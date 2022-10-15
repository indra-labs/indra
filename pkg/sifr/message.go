package sifr

import (
	"crypto/rand"
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
	// Expires are the fingerprints of public keys that the correspondent
	// can now discard as they will not be used again.
	Expires []*schnorr.Pubkey
	// Data is the payload of this message.
	Data []byte
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
