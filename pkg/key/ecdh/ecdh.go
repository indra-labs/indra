package ecdh

import (
	"github.com/Indra-Labs/indra/pkg/key/prv"
	"github.com/Indra-Labs/indra/pkg/key/pub"
	"github.com/Indra-Labs/indra/pkg/sha256"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// Compute computes an elliptic curve diffie hellman shared secret that can be
// decrypted by the holder of the private key matching the public key provided.
func Compute(prv *prv.Key, pub *pub.Key) sha256.Hash {
	return secp256k1.GenerateSharedSecret(
		(*secp256k1.PrivateKey)(prv), (*secp256k1.PublicKey)(pub),
	)
}
