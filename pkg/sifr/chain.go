package sifr

//
// import (
// 	"crypto/aes"
// 	"crypto/cipher"
// 	"fmt"
//
// 	"github.com/Indra-Labs/indra/pkg/schnorr"
// )
//
// type Chain struct {
// 	Recipient   *schnorr.Pubkey
// 	LastPrivkey *schnorr.Privkey
// 	Msg         *Message
// }
//
// type Chains []Chain
//
// // Initiate starts a message cycle with a counterparty.
// func Initiate(msg []byte,
// 	recipient *schnorr.PubkeyBytes) (c *Chain, m *Message, e error) {
//
// 	if recipient == nil {
// 		e = fmt.Errorf("provided public key for recipient nil")
// 		log.E.Chk(e)
// 		return
// 	}
// 	if msg == nil || len(msg) < 1 {
// 		e = fmt.Errorf("empty message")
// 		log.E.Chk(e)
// 		return
// 	}
// 	c = &Chain{}
// 	if c.Recipient, e = recipient.Deserialize(); log.E.Chk(e) {
// 		return
// 	}
// 	if c.LastPrivkey, e = schnorr.GeneratePrivkey(); log.E.Chk(e) {
// 		return
// 	}
// 	secret := c.LastPrivkey.ECDH(c.Recipient)
// 	var ci cipher.Block
// 	if ci, e = aes.NewCipher(secret); log.E.Chk(e) {
// 		return
// 	}
// 	var gcm cipher.AEAD
// 	if gcm, e = cipher.NewGCM(ci); log.E.Chk(e) {
// 		return
// 	}
// 	var sig *schnorr.Signature
// 	sig, e = c.LastPrivkey.Sign(schnorr.SHA256D(msg))
// 	cipherText := make([]byte, len(msg)+32)
// 	cipherText = gcm.Seal(cipherText, c.Msg.Nonce[:], msg, nil)
// 	m = &Message{
// 		To:        recipient.Fingerprint(),
// 		From:      c.LastPrivkey.Pubkey().Serialize(),
// 		Nonce:     GetNonce(),
// 		Message:   cipherText,
// 		Signature: sig.Serialize(),
// 	}
// 	c.Msg = m
// 	return
// }
//
// // Decrypt do dee doo.
// func (c *Chain) Decrypt(msg *Message) (decrypted []byte, e error) {
//
// 	return
// }
