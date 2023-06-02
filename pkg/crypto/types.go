package crypto

import (
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/crypto/sha256"
)

type (
	// Ciphers is a collection of 3 encyrption keys used progressively for a reply
	// message payload.
	Ciphers [3]sha256.Hash

	// Keys is a structure for a pre-formed public/private key set with the public
	// key bytes ready for fast comparisons.
	Keys struct {
		Pub   *Pub
		Bytes PubBytes
		Prv   *Prv
	}

	// Nonces is the collection of 3 encryption nonces associated with the Ciphers.
	Nonces [3]nonce.IV

	// Privs is a collection of 3 private keys used for generating reply headers.
	Privs [3]*Prv

	// Pubs is a collection of 3 public keys used for generating reply headers.
	Pubs [3]*Pub
)

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

// Gen3Nonces generates 3 Initialization Vectors.
func Gen3Nonces() (n Nonces) {
	for i := range n {
		n[i] = nonce.New()
	}
	return
}

// GenCiphers generates a set of 3 Ciphers using Privs and Pubs.
func GenCiphers(prvs Privs, pubs Pubs) (ciphers Ciphers) {
	for i := range prvs {
		ciphers[i] = ComputeSharedSecret(prvs[i], pubs[i])
		log.T.Ln("cipher", i, ciphers[i])
	}
	return
}

// GenNonces generates an arbitrary number of Initialization Vector bytes.
func GenNonces(count int) (n []nonce.IV) {
	n = make([]nonce.IV, count)
	for i := range n {
		n[i] = nonce.New()
	}
	return
}

// GenPingNonces generates  6 Initialization Vector bytes.
func GenPingNonces() (n [6]nonce.IV) {
	for i := range n {
		n[i] = nonce.New()
	}
	return
}

// Generate2Keys generates two Keys.
func Generate2Keys() (one, two *Keys, e error) {
	if one, e = GenerateKeys(); fails(e) {
		return
	}
	if two, e = GenerateKeys(); fails(e) {
		return
	}
	return
}

// GenerateKeys generates one set of Keys.
func GenerateKeys() (k *Keys, e error) {
	k = &Keys{}
	if k.Prv, e = GeneratePrvKey(); fails(e) {
		return
	}
	k.Pub = DerivePub(k.Prv)
	k.Bytes = k.Pub.ToBytes()
	return
}

// GetCipherSet generates a set of Privs and Pubs.
func GetCipherSet() (prvs Privs, pubs Pubs) {
	for i := range prvs {
		prv1, prv2 := GetTwoPrvKeys()
		prvs[i] = prv1
		pubs[i] = DerivePub(prv2)
	}
	return
}

// GetTwoPrvKeys is a helper for tests to generate two new private keys.
func GetTwoPrvKeys() (prv1, prv2 *Prv) {
	var e error
	if prv1, e = GeneratePrvKey(); fails(e) {
		return
	}
	if prv2, e = GeneratePrvKey(); fails(e) {
		return
	}
	return
}

// MakeKeys uses a private key to generate a Keys.
func MakeKeys(pr *Prv) *Keys {
	pubkey := DerivePub(pr)
	return &Keys{pubkey, pubkey.ToBytes(), pr}
}
