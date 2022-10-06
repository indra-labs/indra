package ciph

import (
	"crypto/rand"
)

const NonceSize = 12

type Nonce [NonceSize]byte

// GetNonce reads from a cryptographically secure random number source
func GetNonce() (nonce Nonce, e error) {
	if _, e = rand.Read(nonce[:]); log.E.Chk(e) {
	}
	return
}

//
// import (
// 	"crypto/aes"
// 	"crypto/cipher"
// 	"crypto/rand"
// 	"encoding/hex"
// 	"errors"
// 	"fmt"
// 	"io"
//
// 	"golang.org/x/crypto/argon2"
// )

// // Get returns a GCM cipher given a password string. Note that this cipher
// // must be renewed every 4gb of encrypted data as it is GCM.
// func Get(password []byte) (gcm cipher.AEAD, e error) {
// 	bytes := make([]byte, len(password))
// 	rb := make([]byte, len(password))
// 	copy(bytes, password)
// 	copy(rb, password)
// 	var c cipher.Block
// 	ark := argon2.IDKey(rb, bytes, 1, 64*1024, 4, 32)
// 	if c, e = aes.NewCipher(ark); log.E.Chk(e) {
// 		return
// 	}
// 	if gcm, e = cipher.NewGCM(c); log.E.Chk(e) {
// 	}
// 	for i := range bytes {
// 		bytes[i] = 0
// 		rb[i] = 0
// 	}
// 	return
// }

// // DecryptMessage attempts to decode the received message
// func DecryptMessage(creator string, ciph cipher.AEAD, data []byte) (
// 	msg []byte,
// 	e error,
// ) {
// 	nonceSize := ciph.NonceSize()
// 	msg, e = ciph.Open(nil, data[:nonceSize], data[nonceSize:], nil)
// 	if e != nil {
// 		e = errors.New(fmt.Sprintf("%s %s", creator, e.Error()))
// 	} else {
// 		log.D.Ln("decrypted message", hex.EncodeToString(data[:nonceSize]))
// 	}
// 	return
// }
//
// // EncryptMessage encrypts a message, if the nonce is given it uses that
// // otherwise it generates a new one. If there is no cipher this just returns a
// // message with the given magic prepended.
// func EncryptMessage(
// 	creator string,
// 	ciph cipher.AEAD,
// 	magic []byte,
// 	nonce, data []byte,
// ) (msg []byte, e error) {
// 	if ciph != nil {
// 		if nonce == nil {
// 			nonce, e = GetNonce(ciph)
// 		}
// 		msg = append(
// 			append(magic, nonce...),
// 			ciph.Seal(nil, nonce, data, nil)...,
// 		)
// 	} else {
// 		msg = append(magic, data...)
// 	}
// 	return
// }
