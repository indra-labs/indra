package ciph

import (
	"crypto/aes"
	"crypto/cipher"

	"github.com/Indra-Labs/indra"
	"github.com/Indra-Labs/indra/pkg/key/ecdh"
	"github.com/Indra-Labs/indra/pkg/key/prv"
	"github.com/Indra-Labs/indra/pkg/key/pub"
	"github.com/Indra-Labs/indra/pkg/nonce"
	log2 "github.com/cybriq/proc/pkg/log"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

// GetCipher returns a cipher with a given nonce IV built using ECDH from a
// private and public key.
//
// This function would be used when there is multiple packets to en/decipher
// using a single key derived using ECDH from a private and public key, as using
// Cipher would unnecessarily generate the key for each packet.
func GetCipher(from *prv.Key, to *pub.Key, n nonce.IV) (s cipher.Stream,
	e error) {

	secret := ecdh.Compute(from, to)
	var block cipher.Block
	if block, e = aes.NewCipher(secret); !check(e) {
		s = cipher.NewCTR(block, n[:])
	}
	return
}

// Cipher applies the cipher XORKeyStream onto a slice of bytes. This changes it
// from cleartext to ciphertext and vice versa.
func Cipher(from *prv.Key, to *pub.Key, n nonce.IV, payload []byte) (e error) {
	var stream cipher.Stream
	if stream, e = GetCipher(from, to, n); !check(e) {
		stream.XORKeyStream(payload, payload)
	}
	return
}
