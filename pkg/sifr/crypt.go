package sifr

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"

	"github.com/Indra-Labs/indra/pkg/schnorr"
)

// Crypt is a form of Message that is encrypted. The cipher must
// be conveyed by other means, mainly being ECDH.
type Crypt struct {
	*Nonce
	Message []byte
}

// NewCrypt takes a Message, encrypts it using a secret and returns a Crypt.
func NewCrypt(message *Message, secret schnorr.Hash) (cr *Crypt, e error) {
	if e = secret.Valid(); log.E.Chk(e) {
		return
	}
	var block cipher.Block
	if block, e = aes.NewCipher(secret); log.E.Chk(e) {
		return
	}
	nonce := GetNonce()
	stream := cipher.NewCTR(block, nonce[:])
	msg := message.Serialize()
	stream.XORKeyStream(msg, msg)
	cr = &Crypt{Nonce: nonce, Message: msg}
	return
}

// Serialize converts a Crypt into a generic byte slice.
func (cr *Crypt) Serialize() (b []byte) {
	b = append((*cr.Nonce)[:], cr.Message...)
	return
}

func DeserializeCrypt(message []byte) (cr *Crypt, e error) {
	if len(message) < NonceSize {
		e = fmt.Errorf("message too short to be a Crypt: %d",
			len(message))
		return
	}
	nonce := &Nonce{}
	copy(nonce[:], message[:NonceSize])
	msg := message[NonceSize:]
	cr = &Crypt{Nonce: nonce, Message: msg}
	return
}

func (cr *Crypt) Decrypt(secret schnorr.Hash) (m *Message, e error) {
	if e = secret.Valid(); log.E.Chk(e) {
		return
	}
	var block cipher.Block
	if block, e = aes.NewCipher(secret); log.E.Chk(e) {
		return
	}
	stream := cipher.NewCTR(block, cr.Nonce[:])
	stream.XORKeyStream(cr.Message, cr.Message)
	if m, e = DeserializeMessage(cr.Message); log.E.Chk(e) {
	}
	return
}
