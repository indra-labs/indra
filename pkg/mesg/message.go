package mesg

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/Indra-Labs/indra/pkg/keys"
	"github.com/Indra-Labs/indra/pkg/sha256"
	"github.com/Indra-Labs/indra/pkg/sigs"
)

// Message is simply a wrapper that provides tamper-proofing and
// authentication based on schnorr signatures.
type Message struct {
	Payload   []byte
	Signature sigs.SignatureBytes
}

// New creates a new message with integrity/authentication using a
// schnorr signature.
//
// It is advised to use buffers for the payload with some extra capacity so that
// there is no further allocations in the Serialize function or a following
// encryption step.
func New(payload []byte, prv *keys.Privkey) (m *Message, e error) {
	var sig *sigs.Signature
	if sig, e = prv.Sign(sha256.Hash(payload)); log.E.Chk(e) {
		return
	}
	m = &Message{Payload: payload, Signature: *sig.Serialize()}
	return
}

// Verify checks the signature on the message against a given public key. If
// there is an error with the signature or the message does not validate an
// error is returned.
func (m *Message) Verify(pub *keys.Pubkey) (e error) {
	var sig *sigs.Signature
	if sig, e = m.Signature.Deserialize(); log.E.Chk(e) {
		return
	}
	if !sig.Verify(sha256.Hash(m.Payload), pub) {
		e = errors.New("message did not verify")
	}
	return
}

// Serialize turns the message into a generic byte slice. The message has a 64
// bit length prefix in order to pad to 32 byte segments, maintaining processor
// word alignment and making the whole message align to both hash size and
// AES block size.
func (m *Message) Serialize() (z []byte) {
	// Using 64 bits of length prefix to keep processor word alignment.
	mLen := len(m.Payload)
	padLen := 32 - (mLen % 32)
	mLenBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(mLenBytes, uint64(mLen))
	// Pad the data out to whole divisible amounts of 256 bits with the
	// hash of the first 32 bytes of the message.
	var padBytes []byte
	if padLen > 0 {
		hashDataLen := 32
		if len(m.Payload) < 32 {
			hashDataLen = len(m.Payload)
		}
		padHash := sha256.Hash(m.Payload[:hashDataLen])
		padBytes = padHash[:padLen]
	}
	z = append(append(append(mLenBytes, m.Payload...),
		padBytes...), m.Signature[:]...)
	return
}

// DeserializeMessage accepts a slice of bytes as produced by Serialize and
// returns a Message. This will allocate extra space for the Signature array.
func DeserializeMessage(b []byte) (m *Message, e error) {
	if len(b) < 8 {
		e = fmt.Errorf("message to short, size: %d minimal: %d",
			len(b), 8)
		return
	}
	payloadLen := binary.LittleEndian.Uint64(b[:8])
	padLen := 32 - (payloadLen % 32)
	m = &Message{}
	sigStart := 8 + payloadLen + padLen
	m.Payload = b[8 : payloadLen+8]
	copy(m.Signature[:], b[sigStart:])
	return
}
