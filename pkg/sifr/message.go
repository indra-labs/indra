package sifr

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"sync"

	"github.com/Indra-Labs/indra/pkg/schnorr"
)

const NonceLen = 12

type Nonce [NonceLen]byte

// GetNonce reads from a cryptographically secure random number source
func GetNonce() (nonce *Nonce) {
	if _, e := rand.Read(nonce[:]); log.E.Chk(e) {
	}
	return
}

// Dialog is a data structure for tracking keys used in a message exchange.
type Dialog struct {
	sync.Mutex
	// LastIn is the newest pubkey seen in a received message from the
	// correspondent.
	LastIn *schnorr.Pubkey
	// LastOut is the newest privkey used in an outbound message.
	LastOut *schnorr.Privkey
	// Seen are the keys that have been seen since the last new message sent
	// out to the correspondent.
	Seen []*schnorr.Pubkey
	// Used are the recently used keys that have not been invalidated by the
	// counterparty sending them in the Expires field.
	Used []*schnorr.Privkey
	// UsedFingerprints are 1:1 mapped to Used private keys for fast
	// recognition.
	UsedFingerprints []schnorr.Fingerprint
}

// NewDialog creates a new Dialog for tracking a conversation between two nodes.
// For the initiator, the pubkey is the current one advertised by the
// correspondent, and for a correspondent, this pubkey is from the first one
// appearing in the initial message.
func NewDialog(pub *schnorr.Pubkey) (d *Dialog) {
	d = &Dialog{LastIn: pub}
	return
}

// Outbound is the data structure for constructing an outbound wire format
// message.
type Outbound struct {
	// To is the fingerprint of the pubkey used in the ECDH key exchange.
	To *schnorr.Pubkey
	// From is the pubkey corresponding to the private key used in the ECDH
	// key exchange.
	From *schnorr.Privkey
	// DataLen is the length of the payload of this message.
	Data []byte
	// Expires are the fingerprints of public keys that the correspondent
	// can now discard as they will not be used again.
	Expires []*schnorr.Pubkey
}

const MessageOverhead = schnorr.FingerprintLen + schnorr.PubkeyLen + NonceLen +
	4 + 2 + schnorr.SigLen*3

// Serialize converts a Message into the wire format, signs the payload before
// encrypting and then signs the final packet.
//
// # Binary format
//
// # Field        - Size      - Description
//
// -----------------------------------------------------------------------------
//
// # To           - 8 bytes   - Fingerprint of public key of recipient used in with ECDH for cipher
//
// # From         - 32 bytes  - Public key of sender used with ECDH for cipher
//
// # MessageNonce - 12 bytes  - Cryptographically random nonce for message encryption
//
// # MessageSize  - 4 bytes   - Size of message (up to 4Gb)
//
// # Message      - variable  - Message is signed, signature appended, then encrypted with cipher
//
// # ExpireCount  - 2 bytes   - Number of expired public keys seen prior to dispatch of this message
//
// # Expired      - 8 bytes   - Fingerprint of expired public keys of recipient that have been seen
//
// …                          - repeats per ExpireCount
//
// Signature     - 64 bytes   - Schnorr Signature over entire message data (all previous fields) to prevent tampering
func (m *Outbound) Serialize() (b []byte, e error) {
	// Pre-allocate enough to ensure we have enough bytes for the serialised
	// form of the message.
	b = make([]byte, MessageOverhead+len(m.Data)+
		schnorr.FingerprintLen*len(m.Expires))

	var cursor int
	// To
	toFP := m.To.Fingerprint()
	copy(b[cursor:cursor+len(toFP)], toFP[:])
	cursor += len(toFP)
	// From
	fromKey := m.From.Pubkey().Serialize()
	copy(b[cursor:cursor+len(fromKey)], fromKey[:])
	cursor += len(fromKey)
	// Nonce
	copy(b[cursor:cursor+NonceLen], (*GetNonce())[:])
	cursor += NonceLen
	// Sign message
	var sig *schnorr.Signature
	if sig, e = m.From.Sign(schnorr.SHA256D(m.Data)); log.E.Chk(e) {
		return
	}
	// MessageSize - message length is m.Data plus a signature
	msgLen := len(m.Data) + schnorr.SigLen
	binary.LittleEndian.PutUint32(b[cursor:cursor+4], uint32(msgLen))
	cursor += 4
	// Encrypt data
	buf := make([]byte, msgLen)
	// copy message into buffer with signature appended
	copy(buf, m.Data)
	copy(buf[len(m.Data):], (*sig.Serialize())[:])
	// generate secret
	secret := m.From.ECDH(m.To)
	var ci cipher.Block
	if ci, e = aes.NewCipher(secret); log.E.Chk(e) {
		return
	}
	var gcm cipher.AEAD
	if gcm, e = cipher.NewGCM(ci); log.E.Chk(e) {
	}
	gcm.Seal(buf[:0], b[cursor-NonceLen:cursor], buf, nil)
	// copy encrypted message into buffer
	copy(b[cursor:cursor+len(buf)], buf)
	cursor += msgLen
	// ExpireCount
	binary.LittleEndian.PutUint16(b[cursor:cursor+2], uint16(len(m.Expires)))
	cursor += 2
	// Expired
	for i := range m.Expires {
		x := m.Expires[i].Fingerprint()
		copy(b[cursor:cursor+schnorr.FingerprintLen], x[:])
		cursor += schnorr.FingerprintLen
	}
	if sig, e = m.From.Sign(schnorr.SHA256D(b[:cursor])); log.E.Chk(e) {
		return
	}
	copy(b[cursor:cursor+schnorr.SigLen], sig.Serialize()[:])
	cursor += schnorr.SigLen
	b = b[:cursor]
	return
}

type Inbound struct {
	// To is the fingerprint of the pubkey used in the ECDH key exchange.
	To *schnorr.Fingerprint
	// From is the pubkey corresponding to the private key used in the ECDH
	// key exchange.
	From *schnorr.Pubkey
	// DataLen is the length of the payload of this message.
	Data []byte
	// Expires are the fingerprints of public keys that the correspondent
	// can now discard as they will not be used again.
	Expires []schnorr.Fingerprint
}

func (m *Inbound) Deserialize(b []byte) (e error) {

	return
}
