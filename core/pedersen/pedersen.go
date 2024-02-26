package pedersen

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/cronokirby/saferith"
	"github.com/mr-shifu/mpc-lib/core/math/arith"
	"github.com/mr-shifu/mpc-lib/lib/params"
)

type Error string

const (
	ErrNilFields    Error = "contains nil field"
	ErrSEqualT      Error = "S cannot be equal to T"
	ErrNotValidModN Error = "S and T must be in [1,…,N-1] and coprime to N"
)

func (e Error) Error() string {
	return fmt.Sprintf("pedersen: %s", string(e))
}

type Parameters struct {
	n    *arith.Modulus
	s, t *saferith.Nat
}

// New returns a new set of Pedersen parameters.
// Assumes ValidateParameters(n, s, t) returns nil.
func New(n *arith.Modulus, s, t *saferith.Nat) *Parameters {
	return &Parameters{
		s: s,
		t: t,
		n: n,
	}
}

// ValidateParameters check n, s and t, and returns an error if any of the following is true:
// - n, s, or t is nil.
// - s, t are not in [1, …,n-1].
// - s, t are not coprime to N.
// - s = t.
func ValidateParameters(n *saferith.Modulus, s, t *saferith.Nat) error {
	if n == nil || s == nil || t == nil {
		return ErrNilFields
	}
	// s, t ∈ ℤₙˣ
	if !arith.IsValidNatModN(n, s, t) {
		return ErrNotValidModN
	}
	// s ≡ t
	if _, eq, _ := s.Cmp(t); eq == 1 {
		return ErrSEqualT
	}
	return nil
}

// N = p•q, p ≡ q ≡ 3 mod 4.
func (p Parameters) N() *saferith.Modulus { return p.n.Modulus }

// N, but as an arith modulus, which is sometimes useful
func (p Parameters) NArith() *arith.Modulus { return p.n }

// S = r² mod N.
func (p Parameters) S() *saferith.Nat { return p.s }

// T = Sˡ mod N.
func (p Parameters) T() *saferith.Nat { return p.t }

// Commit computes sˣ tʸ (mod N)
//
// x and y are taken as saferith.Int, because we want to keep these values in secret,
// in general. The commitment produced, on the other hand, hides their values,
// and can be safely shared.
func (p Parameters) Commit(x, y *saferith.Int) *saferith.Nat {
	sx := p.n.ExpI(p.s, x)
	ty := p.n.ExpI(p.t, y)

	result := sx.ModMul(sx, ty, p.n.Modulus)

	return result
}

// Verify returns true if sᵃ tᵇ ≡ S Tᵉ (mod N).
func (p Parameters) Verify(a, b, e *saferith.Int, S, T *saferith.Nat) bool {
	if a == nil || b == nil || S == nil || T == nil || e == nil {
		return false
	}
	nMod := p.n.Modulus
	if !arith.IsValidNatModN(nMod, S, T) {
		return false
	}

	sa := p.n.ExpI(p.s, a)         // sᵃ (mod N)
	tb := p.n.ExpI(p.t, b)         // tᵇ (mod N)
	lhs := sa.ModMul(sa, tb, nMod) // lhs = sᵃ⋅tᵇ (mod N)

	te := p.n.ExpI(T, e)          // Tᵉ (mod N)
	rhs := te.ModMul(te, S, nMod) // rhs = S⋅Tᵉ (mod N)
	return lhs.Eq(rhs) == 1
}

func (p Parameters) MarshalBiinary() ([]byte, error) {
	nb, err := p.n.MarshalBinary()
	if err != nil {
		return nil, err
	}

	sb, err := p.s.MarshalBinary()
	if err != nil {
		return nil, err
	}

	tb, err := p.t.MarshalBinary()
	if err != nil {
		return nil, err
	}

	nlb := make([]byte, 2)
	binary.LittleEndian.PutUint16(nlb, uint16(len(nb)))

	slb := make([]byte, 2)
	binary.LittleEndian.PutUint16(slb, uint16(len(sb)))

	tlb := make([]byte, 2)
	binary.LittleEndian.PutUint16(tlb, uint16(len(tb)))

	buf := make([]byte, 0)
	buf = append(buf, nlb...)
	buf = append(buf, nb...)
	buf = append(buf, slb...)
	buf = append(buf, sb...)
	buf = append(buf, tlb...)
	buf = append(buf, tb...)

	return buf, nil
}

func (p *Parameters) UnmarshalBiinary(data []byte) error {
	nl := binary.LittleEndian.Uint16(data[:2])
	nb := data[2 : 2+nl]

	sl := binary.LittleEndian.Uint16(data[2+nl : 2+nl+2])
	sb := data[2+nl+2 : 2+nl+2+sl]

	tl := binary.LittleEndian.Uint16(data[2+nl+2+sl : 2+nl+2+sl+2])
	tb := data[2+nl+2+sl+2 : 2+nl+2+sl+2+tl]

	n := arith.NewEmptyModulus()
	if err := n.UnmarshalBinary(nb); err != nil {
		return err
	}
	p.n = n

	var s saferith.Nat
	if err := s.UnmarshalBinary(sb); err != nil {
		return err
	}
	p.s = &s

	var t saferith.Nat
	if err := t.UnmarshalBinary(tb); err != nil {
		return err
	}
	p.t = &t

	return nil
}

// WriteTo implements io.WriterTo and should be used within the hash.Hash function.
func (p *Parameters) WriteTo(w io.Writer) (int64, error) {
	if p == nil {
		return 0, io.ErrUnexpectedEOF
	}
	nAll := int64(0)
	buf := make([]byte, params.BytesIntModN)

	// write N, S, T
	for _, i := range []*saferith.Nat{p.n.Nat(), p.s, p.t} {
		i.FillBytes(buf)
		n, err := w.Write(buf)
		nAll += int64(n)
		if err != nil {
			return nAll, err
		}
	}
	return nAll, nil
}

// Domain implements hash.WriterToWithDomain, and separates this type within hash.Hash.
func (Parameters) Domain() string {
	return "Pedersen Parameters"
}
