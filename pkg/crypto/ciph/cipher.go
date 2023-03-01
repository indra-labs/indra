// Package ciph manages encryption ciphers and encrypting blobs of data. Keys
// are generated using ECDH from a public and private secp256k1 combined, as
// well as directly from a 32 byte secret in the form of a static array as used
// in most cryptographic hash function implementations in Go.
package ciph

import (
	"crypto/aes"
	"crypto/cipher"
	
	"git-indra.lan/indra-labs/indra"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/ecdh"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

// GetBlock returns a block cipher with a secret generated from the provided
// keys using ECDH.
func GetBlock(from *prv.Key, to *pub.Key) (block cipher.Block) {
	secret := ecdh.Compute(from, to)
	block, _ = aes.NewCipher(secret[:])
	return
}

// BlockFromHash creates an AES block cipher from an sha256.Hash.
func BlockFromHash(h sha256.Hash) (block cipher.Block) {
	// We can ignore the error because sha256.Hash is a valid key size.
	block, _ = aes.NewCipher(h[:])
	return
}

// Encipher XORs the data with the block stream. This encrypts unencrypted data
// and decrypts encrypted data. If the cipher.Block is nil, it panics (this
// should never happen).
func Encipher(blk cipher.Block, n nonce.IV, b []byte) {
	if blk == nil {
		panic("Encipher called without a block cipher provided")
	} else {
		cipher.NewCTR(blk, n[:]).XORKeyStream(b, b)
	}
}
