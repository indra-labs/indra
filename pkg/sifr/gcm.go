package sifr

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"

	"github.com/Indra-Labs/indra/pkg/schnorr"
)

const NonceLen = 12

type Nonce [NonceLen]byte

type Codec struct {
	priv    *schnorr.Privkey
	sendPub *schnorr.PubkeyBytes
	pub     *schnorr.Pubkey
	schnorr.Hash
	gcm cipher.AEAD
}

// GetNonce reads from a cryptographically secure random number source
func GetNonce() (nonce Nonce) {
	if _, e := rand.Read(nonce[:]); log.E.Chk(e) {
	}
	return
}

// New returns an ECDH encryption Codec for encrypting messages and decoding
// the returned messages.
//
// The private key used here for the initiator of the session should be
// generated newly for each new session, and the public key is the openly
// advertised public key of the recipient, ensuring a stream of messages never
// has a repeating cipher.
//
// Due to the limitations of GCM encryption, this ephemeral sender keypair
// should be rotated after being used on 4Gb of data.
func New(priv *schnorr.Privkey, pub *schnorr.Pubkey) (c *Codec, e error) {
	c = &Codec{priv: priv, sendPub: priv.Pubkey().Serialize(), pub: pub,
		Hash: priv.ECDH(pub)}
	var ci cipher.Block
	if ci, e = aes.NewCipher(c.Hash); log.E.Chk(e) {
		return
	}
	if c.gcm, e = cipher.NewGCM(ci); log.E.Chk(e) {
	}
	return
}

func (c *Codec) Encrypt(message []byte) (msg *Message, e error) {
	msg = &Message{
		Pubkey:  *c.sendPub,
		Nonce:   GetNonce(),
		Message: make([]byte, len(message)+16),
	}
	c.gcm.Seal(msg.Message, msg.Nonce[:], message, nil)
	// get hash of cleartext message for signature
	cleartextHash := schnorr.SHA256D(message)
	// erase cleartext message
	for i := range message {
		message[i] = 0
	}
	var sig *schnorr.Signature
	if sig, e = c.priv.Sign(cleartextHash); log.E.Chk(e) {
		return
	}
	msg.Signature = *sig.Serialize()
	return
}

func (c *Codec) Decrypt(msg *Message) (payload []byte, e error) {
	if *c.sendPub != msg.Pubkey {
		var pub *schnorr.Pubkey
		pub, e = (&msg.Pubkey).Deserialize()
		e = fmt.Errorf("message does not appear to be for this cipher"+
			"pubkey %v expected, %v received, self-decryption?",
			c.sendPub.Fingerprint(), pub.Fingerprint(),
		)
		return
	}
	if payload, e = c.gcm.Open(nil, msg.Nonce[:], msg.Message, nil); log.E.Chk(e) {
		return
	}
	hash := schnorr.SHA256D(payload)
	var sig *schnorr.Signature
	if sig, e = msg.Signature.Deserialize(); log.E.Chk(e) {
		return
	}
	if !sig.Verify(hash, c.pub) {
		e = fmt.Errorf("signature did not correspond to payload, or" +
			" was not made by same key as used to generate secret")
	}
	return
}
