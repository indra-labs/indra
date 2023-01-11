// Package ecdh is provides a function to take a secp256k1 public and private
// key pair to generate a shared secret that the corresponding opposite half
// from the counterparty can use to generate the same cipher.
package ecdh

import (
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/indra-labs/indra/pkg/key/prv"
	"github.com/indra-labs/indra/pkg/key/pub"
	"github.com/indra-labs/indra/pkg/sha256"
)

// Compute computes an Elliptic Curve Diffie-Hellman shared secret that can be
// decrypted by the holder of the private key matching the public key provided.
func Compute(prv *prv.Key, pub *pub.Key) sha256.Hash {
	return sha256.Single(
		secp256k1.GenerateSharedSecret(
			(*secp256k1.PrivateKey)(prv), (*secp256k1.PublicKey)(pub),
		),
	)
}
