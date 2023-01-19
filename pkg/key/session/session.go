package session

import (
	"github.com/Indra-Labs/indra/pkg/key/address"
)

// A Session keeps track of a connection session. It specifically maintains the
// account of available bandwidth allocation before it needs to be recharged
// with new credit, and the current state of the encryption.
type Session struct {
	Remaining    uint64
	SendEntry    *address.SendEntry
	ReceiveEntry *address.ReceiveEntry
}

type Sessions []*Session

// New creates a new Session.
//
// Purchasing a session the seller returns a token, based on a requested data
// allocation,
func New(rem uint64, se *address.SendEntry, re *address.ReceiveEntry) *Session {
	return &Session{
		Remaining:    rem,
		SendEntry:    se,
		ReceiveEntry: re,
	}
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
