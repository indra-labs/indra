package client

import (
	"github.com/Indra-Labs/indra/pkg/key/address"
	"github.com/Indra-Labs/indra/pkg/key/signer"
	"github.com/Indra-Labs/indra/pkg/nonce"
)

// A Session keeps track of a connection session. It specifically maintains the
// account of available bandwidth allocation before it needs to be recharged
// with new credit, and the current state of the encryption.
type Session struct {
	nonce.ID
	Remaining    uint64
	SendEntry    *address.SendEntry
	ReceiveEntry *address.ReceiveEntry
	KeyRoller    *signer.KeySet
}

type Sessions []*Session

func (s Sessions) Len() int {
	return len(s)
}

func (s Sessions) Add(se *Session) Sessions {
	return append(s, se)
}

func (s Sessions) Delete(se *Session) Sessions {
	for i := range s {
		if s[i] == se {
			return append(s[:i], s[i:]...)
		}
	}
	return s
}

func (s Sessions) Find(t nonce.ID) (se *Session) {
	for i := range s {
		if s[i].ID == t {
			se = s[i]
		}
	}
	return
}

// NewSession creates a new Session.
//
// Purchasing a session the seller returns a token, based on a requested data
// allocation,
func NewSession(id nonce.ID, rem uint64, se *address.SendEntry,
	re *address.ReceiveEntry, kr *signer.KeySet) (s *Session) {

	s = &Session{
		ID:           id,
		Remaining:    rem,
		SendEntry:    se,
		ReceiveEntry: re,
		KeyRoller:    kr,
	}
	return
}

// AddBytes adds to the Remaining counter, used when new data allowance has been
// purchased.
func (s *Session) AddBytes(b uint64) {
	s.Remaining += b
}

// SubtractBytes reduces the amount Remaining, if the requested amount would put
// the total below zero it returns false, signalling that new data allowance
// needs to be purchased before any further messages can be sent.
func (s *Session) SubtractBytes(b uint64) bool {
	if s.Remaining < b {
		return false
	}
	s.Remaining -= b
	return true
}

func (s *Session) SetSendEntry(se *address.SendEntry) {
	s.SendEntry = se
}

func (s *Session) SetReceiveEntry(re *address.ReceiveEntry) {
	s.ReceiveEntry = re
}
