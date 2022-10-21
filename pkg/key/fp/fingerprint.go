package fp

import (
	"github.com/Indra-Labs/indra/pkg/key/pub"
)

const (
	Len         = 8
	ReceiverLen = 12
)

type (
	Key      []byte
	Receiver []byte
)

// New empty public Key fingerprint.
func New() Key { return make(Key, Len) }

// NewReceiver makes a new empty Receiver public key fingerprint.
func NewReceiver() Receiver { return make(Receiver, ReceiverLen) }

// Get creates a slice of fingerprints from a set of public keys.
func Get(keys ...*pub.Key) (fps []Key) {
	for i := range keys {
		fps = append(fps, keys[i].ToBytes().Fingerprint())
	}
	return
}
