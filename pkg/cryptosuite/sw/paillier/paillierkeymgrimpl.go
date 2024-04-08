package paillier

import (
	"encoding/hex"
	"errors"

	"github.com/cronokirby/saferith"
	comm_paillier "github.com/mr-shifu/mpc-lib/pkg/common/cryptosuite/paillier"
	"github.com/mr-shifu/mpc-lib/pkg/common/keystore"
	"github.com/mr-shifu/mpc-lib/pkg/common/keyopts"

	pailliercore "github.com/mr-shifu/mpc-lib/core/paillier"
	"github.com/mr-shifu/mpc-lib/core/pool"
)

type PaillierKeyManager struct {
	pl       *pool.Pool
	keystore keystore.Keystore
}

func NewPaillierKeyManager(store keystore.Keystore, pl *pool.Pool) *PaillierKeyManager {
	return &PaillierKeyManager{
		keystore: store,
	}
}

// GenerateKey generates a new Paillier key pair.
func (mgr *PaillierKeyManager) GenerateKey(opts keyopts.Options) (comm_paillier.PaillierKey, error) {
	// generate a new Paillier key pair
	pk, sk := pailliercore.KeyGen(mgr.pl)
	key := PaillierKey{sk, pk}

	// get binary encoded of secret key params (P, Q)
	encoded, err := key.Bytes()
	if err != nil {
		return PaillierKey{}, err
	}

	// derive SKI from N param of public key
	ski := key.SKI()
	keyID := hex.EncodeToString(ski)

	// store the key to the keystore with keyID
	if err := mgr.keystore.Import(keyID, encoded, opts); err != nil {
		return PaillierKey{}, err
	}

	return key, nil
}

// GetKey returns a Paillier key by its SKI.
func (mgr *PaillierKeyManager) GetKey(opts keyopts.Options) (comm_paillier.PaillierKey, error) {
	// get the key from the keystore
	// keyID := hex.EncodeToString(ski)
	decoded, err := mgr.keystore.Get(opts)
	if err != nil {
		return PaillierKey{}, nil
	}

	// decode the key from the keystore
	key, err := fromBytes(decoded)
	if err != nil {
		return PaillierKey{}, nil
	}

	return key, nil
}

// ImportKey imports a Paillier key from its byte representation.
func (mgr *PaillierKeyManager) ImportKey(raw interface{}, opts keyopts.Options) (comm_paillier.PaillierKey, error) {
	var err error
	var key PaillierKey

	switch raw := raw.(type) {
	case []byte:
		key, err = fromBytes(raw)
		if err != nil {
			return PaillierKey{}, err
		}
	case PaillierKey:
		key = raw
	}

	if err := pailliercore.ValidateN(key.ParamN()); err != nil {
		return nil, errors.New("invalid Paillier key")
	}

	// encode the key into binary
	kb, err := key.Bytes()
	if err != nil {
		return nil, err
	}

	// get SKI from key
	ski := key.SKI()
	keyID := hex.EncodeToString(ski)

	// store the key to the keystore with keyID
	if err := mgr.keystore.Import(keyID, kb, opts); err != nil {
		return nil, errors.New("failed to import key")
	}

	return key, nil
}

// Encrypt returns the encryption of `message` as ciphertext and nonce generated by function.
func (mgr *PaillierKeyManager) Encode(m *saferith.Int, opts keyopts.Options) (*pailliercore.Ciphertext, *saferith.Nat) {
	key, err := mgr.GetKey(opts)
	if err != nil {
		return nil, nil
	}

	return key.Encode(m)
}

// EncryptWithNonce returns the encryption of `message` as ciphertext and nonce passed to function.
func (mgr *PaillierKeyManager) EncWithNonce(m *saferith.Int, nonce *saferith.Nat, opts keyopts.Options) *pailliercore.Ciphertext {
	key, err := mgr.GetKey(opts)
	if err != nil {
		return nil
	}

	return key.EncWithNonce(m, nonce)
}

// Decrypt returns the decryption of `ct` as ciphertext.
func (mgr *PaillierKeyManager) Decode(ct *pailliercore.Ciphertext, opts keyopts.Options) (*saferith.Int, error) {
	key, err := mgr.GetKey(opts)
	if err != nil {
		return nil, err
	}

	return key.Decode(ct)
}

// DecryptWithNonce returns the decryption of `ct` as ciphertext and nonce.
func (mgr *PaillierKeyManager) DecodeWithNonce(ct *pailliercore.Ciphertext, opts keyopts.Options) (*saferith.Int, *saferith.Nat, error) {
	key, err := mgr.GetKey(opts)
	if err != nil {
		return nil, nil, err
	}

	return key.DecodeWithNonce(ct)
}

// ValidateCiphertexts returns true if all ciphertexts are valid.
func (mgr *PaillierKeyManager) ValidateCiphertexts(opts keyopts.Options, cts ...*pailliercore.Ciphertext) (bool, error) {
	key, err := mgr.GetKey(opts)
	if err != nil {
		return false, err
	}

	return key.ValidateCiphertexts(cts...), nil
}
