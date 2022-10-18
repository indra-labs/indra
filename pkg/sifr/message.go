package sifr

import (
	"errors"
	"fmt"

	"github.com/Indra-Labs/indra/pkg/schnorr"
	"github.com/Indra-Labs/indra/pkg/sha256"
)

// Message is simply a wrapper that provides tamper-proofing and
// authentication based on schnorr signatures.
type Message struct {
	Payload   []byte
	Signature schnorr.SignatureBytes
}

// NewMessage creates a new message with integrity/authentication using a
// schnorr signature.
//
// It is advised to use buffers for the payload with some extra capacity so that
// there is no further allocations in the Serialize function or a following
// encryption step.
func NewMessage(payload []byte, prv *schnorr.Privkey) (m *Message, e error) {
	var sig *schnorr.Signature
	if sig, e = prv.Sign(sha256.Hash(payload)); log.E.Chk(e) {
		return
	}
	m = &Message{Payload: payload, Signature: *sig.Serialize()}
	return
}

// Verify checks the signature on the message against a given public key. If
// there is an error with the signature or the message does not validate an
// error is returned.
func (m *Message) Verify(pub *schnorr.Pubkey) (e error) {
	var sig *schnorr.Signature
	if sig, e = m.Signature.Deserialize(); log.E.Chk(e) {
		return
	}
	if !sig.Verify(sha256.Hash(m.Payload), pub) {
		e = errors.New("message did not verify")
	}
	return
}

// Serialize turns the message into a generic byte slice.
func (m *Message) Serialize() (z []byte) {
	z = append(m.Payload, m.Signature[:]...)
	return
}

// DeserializeMessage accepts a slice of bytes as produced by Serialize and
// returns a Message. This will allocate extra space for the Signature array.
func DeserializeMessage(b []byte) (m *Message, e error) {
	if len(b) < schnorr.SigLen {
		e = fmt.Errorf("message to short, size: %d minimal: %d",
			len(b), schnorr.SigLen)
		return
	}
	m = &Message{}
	sigStart := len(b) - schnorr.SigLen
	m.Payload = b[:sigStart]
	copy(m.Signature[:], b[sigStart:])
	return
}
