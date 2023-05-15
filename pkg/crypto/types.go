package crypto

import (
	"testing"
	
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
)

type Ciphers [3]sha256.Hash
type Nonces [3]nonce.IV
type Privs [3]*Prv
type Pubs [3]*Pub
type Keys struct {
	Pub   *Pub
	Bytes PubBytes
	Prv   *Prv
}

func GenerateKeys() (k *Keys, e error) {
	k = &Keys{}
	if k.Prv, e = GeneratePrvKey(); fails(e) {
		return
	}
	k.Pub = DerivePub(k.Prv)
	k.Bytes = k.Pub.ToBytes()
	return
}

func Generate2Keys() (one, two *Keys, e error) {
	if one, e = GenerateKeys(); fails(e) {
		return
	}
	if two, e = GenerateKeys(); fails(e) {
		return
	}
	return
}

func MakeKeys(pr *Prv) *Keys {
	pubkey := DerivePub(pr)
	return &Keys{pubkey, pubkey.ToBytes(), pr}
}

// ComputeSharedSecret computes an Elliptic Curve Diffie-Hellman shared secret
// that can be decrypted by the holder of the private key matching the public
// key provided.
func ComputeSharedSecret(prv *Prv, pub *Pub) sha256.Hash {
	return sha256.Single(
		secp256k1.GenerateSharedSecret(
			(*secp256k1.PrivateKey)(prv), (*secp256k1.PublicKey)(pub),
		),
	)
}

func GenCiphers(prvs Privs, pubs Pubs) (ciphers Ciphers) {
	for i := range prvs {
		ciphers[i] = ComputeSharedSecret(prvs[i], pubs[i])
		log.T.Ln("cipher", i, ciphers[i].Based32String())
	}
	return
}

func GenNonces(count int) (n []nonce.IV) {
	n = make([]nonce.IV, count)
	for i := range n {
		n[i] = nonce.New()
	}
	return
}

func GetTwoPrvKeys(t *testing.T) (prv1, prv2 *Prv) {
	var e error
	if prv1, e = GeneratePrvKey(); fails(e) {
		t.FailNow()
	}
	if prv2, e = GeneratePrvKey(); fails(e) {
		t.FailNow()
	}
	return
}

func GetCipherSet(t *testing.T) (prvs Privs, pubs Pubs) {
	for i := range prvs {
		prv1, prv2 := GetTwoPrvKeys(t)
		prvs[i] = prv1
		pubs[i] = DerivePub(prv2)
	}
	return
}

func Gen3Nonces() (n Nonces) {
	for i := range n {
		n[i] = nonce.New()
	}
	return
}

func GenPingNonces() (n [6]nonce.IV) {
	for i := range n {
		n[i] = nonce.New()
	}
	return
}
