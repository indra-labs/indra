package crypto

import (
	"fmt"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
	"github.com/indra-labs/indra/pkg/crypto/sha256"
	"github.com/libp2p/go-libp2p/core/crypto"
	crypto_pb "github.com/libp2p/go-libp2p/core/crypto/pb"
)

// Equals is an implementation of the libp2p crypto.Key interface, allowing the
// Indra keys to be used by libp2p as peer identity keys.
func (p *Prv) Equals(key crypto.Key) (eq bool) {
	var e error
	var rawA, rawB []byte
	if rawA, e = key.Raw(); fails(e) {
		return
	}
	if rawB, e = p.Raw(); fails(e) {
		return
	}
	if len(rawA) != len(rawB) {
		return
	}
	for i := range rawA {
		if rawA[i] != rawB[i] {
			for j := range rawA {
				rawA[j], rawB[j] = 0, 0
			}
			return
		}
	}
	return true
}

// GetPublic derives the public key matching a private key, an implementation of
// the libp2p crypto.PrivKey interface, allowing the Indra keys to be used by libp2p
// as peer identity keys.
func (p *Prv) GetPublic() crypto.PubKey {
	if p == nil {
		return nil
	}
	return DerivePub(p)
}

// Raw is an implementation of the libp2p crypto.Key interface, allowing the
// Indra keys to be used by libp2p as peer identity keys.
func (p *Prv) Raw() ([]byte, error) {
	b := p.ToBytes()
	return b[:], nil
}

// Sign is an implementation of the libp2p crypto.PrivKey interface, allowing the
// Indra keys to be used by libp2p as peer identity keys.
func (p *Prv) Sign(bytes []byte) ([]byte, error) {
	hash := sha256.Single(bytes)
	s := ecdsa.Sign((*secp256k1.PrivateKey)(p), hash[:])
	return s.Serialize(), nil
}

// Type is an implementation of the libp2p crypto.Key interface, allowing the
// Indra keys to be used by libp2p as peer identity keys.
func (p *Prv) Type() crypto_pb.KeyType {
	return crypto_pb.KeyType_Secp256k1
}

// Equals compares two public keys and returns true if they match.
func (k *Pub) Equals(key crypto.Key) (eq bool) {
	var e error
	var rawA, rawB []byte
	if rawA, e = key.Raw(); fails(e) {
		return
	}
	if rawB, e = k.Raw(); fails(e) {
		return
	}
	if len(rawA) != len(rawB) {
		return
	}
	for i := range rawA {
		if rawA[i] != rawB[i] {
			for j := range rawA {
				rawA[j], rawB[j] = 0, 0
			}
			return
		}
	}
	return true
}

// Raw is an implementation of the libp2p crypto.Key interface, allowing the
// Indra keys to be used by libp2p as peer identity keys.
func (k *Pub) Raw() ([]byte, error) {
	b := k.ToBytes()
	return b[:], nil
}

// Type is an implementation of the libp2p crypto.Key interface, allowing the
// Indra keys to be used by libp2p as peer identity keys.
func (k *Pub) Type() crypto_pb.KeyType {
	return crypto_pb.KeyType_Secp256k1
}

// Verify is an implementation of the libp2p crypto.PkbKey interface, allowing the
// Indra keys to be used by libp2p as peer identity keys.
func (k *Pub) Verify(data []byte, sigBytes []byte) (is bool,
	e error) {

	var s SigBytes
	if len(sigBytes) != len(s) {
		return false, fmt.Errorf("length mismatch")
	}
	copy(s[:], sigBytes[:])
	hash := sha256.Single(data)
	var pk *Pub
	if pk, e = s.Recover(hash); fails(e) {
		return false, e
	}
	return pk.ToBytes().Equals(k.ToBytes()), nil
}
