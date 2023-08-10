// Package ciph manages encryption ciphers and encrypting blobs of data. Keys
// are generated using ECDH from a public and private secp256k1 combined, as
// well as directly from a 32 byte secret in the form of a static array as used
// in most cryptographic hash function implementations in Go.
package ciph

import (
	"crypto/aes"
	"crypto/cipher"
	"git.indra-labs.org/dev/ind/pkg/crypto"
	"git.indra-labs.org/dev/ind/pkg/crypto/nonce"
	"git.indra-labs.org/dev/ind/pkg/crypto/sha256"
	log2 "git.indra-labs.org/dev/ind/pkg/proc/log"
)

var (
	log   = log2.GetLogger()
	fails = log.E.Chk
)

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

// GetBlock returns a block cipher with a secret generated from the provided
// keys using ECDH.
func GetBlock(from *crypto.Prv, to *crypto.Pub, note string) (block cipher.Block) {
	secret := crypto.ComputeSharedSecret(from, to)
	// fb := from.ToBytes()
	// log.T.Ln(note, "secret", color.Red.Sprint(enc(secret[:])[:52]), "<-",
	// 	color.Blue.Sprint(enc(fb[:])[:52]), "+", to.ToBased32())
	block, _ = aes.NewCipher(secret[:])
	return
}
