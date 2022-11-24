// Package ecdh is provides a function to take a secp256k1 public and private
// key pair to generate a shared secret that the corresponding opposite half
// from the counterparty can use to generate the same cipher.
package ecdh

import (
	"github.com/Indra-Labs/indra/pkg/blake3"
	"github.com/Indra-Labs/indra/pkg/key/prv"
	"github.com/Indra-Labs/indra/pkg/key/pub"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// Compute computes an elliptic curve diffie hellman shared secret that can be
// decrypted by the holder of the private key matching the public key provided.
func Compute(prv *prv.Key, pub *pub.Key) blake3.Hash {
	return secp256k1.GenerateSharedSecret(
		(*secp256k1.PrivateKey)(prv), (*secp256k1.PublicKey)(pub),
	)
}
