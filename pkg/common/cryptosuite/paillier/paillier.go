package paillier

import (
	"math/big"

	"github.com/cronokirby/saferith"
	"github.com/mr-shifu/mpc-lib/pkg/common/cryptosuite/hash"
	"github.com/mr-shifu/mpc-lib/core/math/arith"
	pailliercore "github.com/mr-shifu/mpc-lib/core/paillier"
	"github.com/mr-shifu/mpc-lib/core/pool"
	zkfac "github.com/mr-shifu/mpc-lib/core/zk/fac"
	zkmod "github.com/mr-shifu/mpc-lib/core/zk/mod"
	"github.com/mr-shifu/mpc-lib/pkg/common/cryptosuite/pedersen"
)

type PaillierKey interface {
	// Bytes returns the byte representation of the key.
	Bytes() ([]byte, error)

	// SKI returns the serialized key identifier.
	SKI() []byte

	// Private returns true if the key is private.
	Private() bool

	// PublicKey returns the corresponding public key part of Elgamal Key.
	PublicKey() PaillierKey

	PublicKeyRaw() *pailliercore.PublicKey

	// Modulus returns an arith.Modulus for N.
	Modulus() *arith.Modulus

	// ParamN returns the public key modulus N.
	ParamN() *saferith.Modulus

	// Encrypt returns the encryption of `message` as ciphertext and nonce generated by function.
	Encode(m *saferith.Int) (*pailliercore.Ciphertext, *saferith.Nat)

	// EncryptWithNonce returns the encryption of `message` as ciphertext and nonce passed to function.
	EncWithNonce(m *saferith.Int, nonce *saferith.Nat) *pailliercore.Ciphertext

	// Decrypt returns the decryption of `ct` as ciphertext.
	Decode(ct *pailliercore.Ciphertext) (*saferith.Int, error)

	// DecryptWithNonce returns the decryption of `ct` as ciphertext and nonce.
	DecodeWithNonce(ct *pailliercore.Ciphertext) (*saferith.Int, *saferith.Nat, error)

	// Sample returns a random number in [1, N-1] and its corresponding big.Int.
	Sample(t *saferith.Nat) (*saferith.Nat, *big.Int)

	// Derive Pedersen Key from Paillier Key prime factors
	DerivePedersenKey() (pedersen.PedersenKey, error)

	// ValidateCiphertexts returns true if all ciphertexts are valid.
	ValidateCiphertexts(cts ...*pailliercore.Ciphertext) bool

	// NewZKModProof returns a new ZKMod proof of paillier key params.
	NewZKModProof(hash hash.Hash, pl *pool.Pool) *zkmod.Proof

	// VerifyZKMod verifies a ZKMod proof of paillier key params.
	VerifyZKMod(p *zkmod.Proof, hash hash.Hash, pl *pool.Pool) bool

	// NewZKFACProof returns a new ZKFAC proof of paillier key params N and other party's pedersen params.
	NewZKFACProof(hash hash.Hash, public zkfac.Public) *zkfac.Proof

	// VerifyZKFAC verifies a ZKFAC proof of paillier key params.
	VerifyZKFAC(p *zkfac.Proof, public zkfac.Public, hash hash.Hash) bool
}

type PaillierKeyManager interface {
	// GenerateKey generates a new Paillier key pair.
	GenerateKey() (PaillierKey, error)

	// GetKey returns a Paillier key by its SKI.
	GetKey(ski []byte) (PaillierKey, error)

	// ImportKey imports a Paillier key from its byte representation.
	ImportKey(key PaillierKey) (PaillierKey, error)

	// Encrypt returns the encryption of `message` as ciphertext and nonce generated by function.
	Encode(ski []byte, m *saferith.Int) (*pailliercore.Ciphertext, *saferith.Nat)

	// EncryptWithNonce returns the encryption of `message` as ciphertext and nonce passed to function.
	EncWithNonce(ski []byte, m *saferith.Int, nonce *saferith.Nat) *pailliercore.Ciphertext

	// Decrypt returns the decryption of `ct` as ciphertext.
	Decode(ski []byte, ct *pailliercore.Ciphertext) (*saferith.Int, error)

	// DecryptWithNonce returns the decryption of `ct` as ciphertext and nonce.
	DecodeWithNonce(ski []byte, ct *pailliercore.Ciphertext) (*saferith.Int, *saferith.Nat, error)

	// ValidateCiphertexts returns true if all ciphertexts are valid.
	ValidateCiphertexts(ski []byte, cts ...*pailliercore.Ciphertext) (bool, error)
}
