package sifr

import (
	"fmt"

	"github.com/Indra-Labs/indra/pkg/schnorr"
)

const MessageOverhead = schnorr.PubkeyLen + NonceLen + schnorr.SigLen

type Message struct {
	// Pubkey corresponds to the private key generated for
	// the message/session, to be combined with recipient's
	// advertised Pubkey's Privkey for the encryption using
	// ECDH.
	Pubkey schnorr.PubkeyBytes
	// Nonce is a 12 byte cryptographically random value used
	// to provide entropy to the cipher.
	Nonce Nonce
	// Message is the payload data.
	Message []byte
	// Signature corresponds to decrypted message, it must
	// match the Pubkey above and the hash of the decrypted
	// message.
	Signature schnorr.SignatureBytes
}

func (msg *Message) ToBytes() (bytes []byte) {
	bytes = make([]byte, MessageOverhead, len(msg.Message))
	var cursor int
	copy(bytes[:schnorr.PubkeyLen], msg.Pubkey[:])
	cursor += schnorr.PubkeyLen
	copy(bytes[cursor:cursor+NonceLen], msg.Nonce[:])
	cursor += NonceLen
	copy(bytes[cursor:cursor+len(msg.Message)], msg.Message)
	cursor += len(msg.Message)
	copy(bytes[cursor:], msg.Signature[:])
	return
}

func FromBytes(msg []byte) (message *Message, e error) {
	msgLen := len(msg)
	if msgLen < MessageOverhead {
		e = fmt.Errorf("message too short, minimum size: %s bytes",
			len(msg))
		log.E.Ln(e)
		return
	}
	payloadLen := msgLen - MessageOverhead
	var cursor int
	copy(message.Pubkey[:], msg[cursor:cursor+schnorr.PubkeyLen])
	cursor += schnorr.PubkeyLen
	copy(message.Nonce[:], msg[cursor:cursor+NonceLen])
	cursor += NonceLen
	// we can avoid copying the payload
	message.Message = msg[cursor : cursor+payloadLen]
	cursor += payloadLen
	copy(message.Signature[:], msg[cursor:msgLen])
	return
}
