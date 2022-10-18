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

func DeserializeCrypt(b []byte) (cr *Crypt, e error) {
	if len(b) < NonceSize {
		e = fmt.Errorf("message too short to be a Crypt: %d", len(b))
		return
	}
	cr = &Crypt{Message: b[NonceSize:]}
	copy(cr.Nonce[:], b[:NonceSize])
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

func DecryptMessage(secret schnorr.Hash, message []byte) (out *Message,
	e error) {

	if len(message) < NonceSize+schnorr.SigLen {
		e = fmt.Errorf("message shorter than nonce + signature, "+
			"minimum %d got %d",
			NonceSize+schnorr.SigLen, len(message))
		log.E.Chk(e)
		return
	}
	if e = secret.Valid(); log.E.Chk(e) {
		return
	}
	nonce := &Nonce{}
	copy(nonce[:], message[:NonceSize])
	msg := message[NonceSize:]
	em := &Crypt{Nonce: nonce, Message: msg}
	var block cipher.Block
	if block, e = aes.NewCipher(secret); log.E.Chk(e) {
		return
	}
	stream := cipher.NewCTR(block, nonce[:])
	stream.XORKeyStream(em.Message, em.Message)
	sigStart := len(msg) - schnorr.SigLen
	m, s := em.Message[:sigStart], em.Message[sigStart:]
	var sig schnorr.SignatureBytes
	copy(sig[:], s)
	out = &Message{Payload: m, Signature: sig}
	return
}
