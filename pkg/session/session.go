// Package session provides a mechanism for keeping track of existing sessions
// with Indra relays.
//
// These sessions consist of the current state of the message encryption scheme
// and the account of remaining data allocation on the session.
package session

import (
	"github.com/Indra-Labs/indra"
	"github.com/Indra-Labs/indra/pkg/key/address"
	"github.com/Indra-Labs/indra/pkg/nonce"
	log2 "github.com/cybriq/proc/pkg/log"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

// A Session keeps track of a connection session. It specifically maintains the
// account of available bandwidth allocation before it needs to be recharged
// with new credit, and the current state of the encryption.
type Session struct {
	nonce.ID
	Remaining    uint64
	SendEntry    *address.SendEntry
	ReceiveEntry *address.ReceiveEntry
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

// New creates a new Session.
//
// Purchasing a session the seller returns a token, based on a requested data
// allocation,
func New(rem uint64, se *address.SendEntry, re *address.ReceiveEntry) (s *Session) {
	s = &Session{
		ID:           nonce.NewID(),
		Remaining:    rem,
		SendEntry:    se,
		ReceiveEntry: re,
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
