// Package ciph manages encryption ciphers and encrypting blobs of data.
package ciph

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"reflect"
	"unsafe"

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

// GetBlock returns a block cipher with a secret generated from the provided
// keys using ECDH.
func GetBlock(from *prv.Key, to *pub.Key) (block cipher.Block, e error) {
	secret := ecdh.Compute(from, to)
	if block, e = aes.NewCipher(secret); !check(e) {
	}
	return
}

// Encipher XORs the data with the block stream. This encrypts unencrypted data
// and decrypts encrypted data.
func Encipher(blk cipher.Block, n nonce.IV, b []byte) {
	if blk != nil {
		cipher.NewCTR(blk, n).XORKeyStream(b, b)
	}
}

const SecretLength = 32

func CombineCiphers(secrets ...[]byte) (combined []byte, e error) {
	// All ciphers must be the same length, 32 bytes for AES-CTR 256
	for i := range secrets {
		if len(secrets[i]) != SecretLength {
			e = fmt.Errorf("unable to combine ciphers, cipher %d is"+
				" length %d, expected %d",
				i, len(secrets[i]), SecretLength)
		}
	}
	combined = make([]byte, SecretLength)
	for i := range combined {
		for j := range secrets {
			combined[i] ^= secrets[j][i]
		}
	}
	return
}

type CipherKeys struct {
	From *prv.Key
	To   *pub.Key
}

func GenerateCompositeCipher(ck ...*CipherKeys) (secret []byte, e error) {
	// var secret64 []uint64
	var secrets [][]byte
	for i := range ck {
		var sec []byte
		sec = ecdh.Compute(ck[i].From, ck[i].To)
		secrets = append(secrets, sec)
		// sec64 := touint64(sec)
		// if i == 0 {
		// 	secret64 = sec64
		// } else {
		// 	xor64(secret64, sec64)
		// }
	}
	return CombineCiphers(secrets...)
}

// touint64 assumes len(x)%8 == 0
func touint64(x []byte) []uint64 {
	xx := make([]uint64, 0, 0)
	hdrp := (*reflect.SliceHeader)(unsafe.Pointer(&xx))
	hdrp.Data = (*reflect.SliceHeader)(unsafe.Pointer(&x)).Data
	hdrp.Len = len(x) / 8
	hdrp.Cap = len(x) / 8
	return xx
}

func tobytes(x []uint64) []byte {
	xx := make([]byte, 0, 0)
	hdrp := (*reflect.SliceHeader)(unsafe.Pointer(&xx))
	hdrp.Data = (*reflect.SliceHeader)(unsafe.Pointer(&x)).Data
	hdrp.Len = len(x) * 8
	hdrp.Cap = len(x) * 8
	return xx
}

func xor64(a, b []uint64) {
	for i := range a {
		a[i] ^= b[i]
	}
}
