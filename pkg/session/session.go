package session

import (
	"time"

	"github.com/indra-labs/indra"
	"github.com/indra-labs/indra/pkg/key/prv"
	"github.com/indra-labs/indra/pkg/key/pub"
	log2 "github.com/indra-labs/indra/pkg/log"
	"github.com/indra-labs/indra/pkg/node"
	"github.com/indra-labs/indra/pkg/nonce"
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
	*node.Node
	Remaining                 uint64
	HeaderPub, PayloadPub     *pub.Key
	HeaderBytes, PayloadBytes pub.Bytes
	HeaderPrv, PayloadPrv     *prv.Key
	Depth                     int8
	Deadline                  time.Time
}

// New creates a new Session.
//
// Purchasing a session the seller returns a token, based on a requested data
// allocation.
func New(id nonce.ID, rem uint64, deadline time.Duration) (s *Session) {

	var e error
	var hdrPrv, pldPrv *prv.Key
	if hdrPrv, e = prv.GenerateKey(); check(e) {
	}
	hdrPub := pub.Derive(hdrPrv)
	// hdrSend := address.NewSendEntry(hdrPub)
	if pldPrv, e = prv.GenerateKey(); check(e) {
	}
	pldPub := pub.Derive(pldPrv)
	// pldSend := address.NewSendEntry(pldPub)

	s = &Session{
		ID:           id,
		Remaining:    rem,
		HeaderPub:    hdrPub,
		HeaderBytes:  hdrPub.ToBytes(),
		PayloadPub:   pldPub,
		PayloadBytes: pldPub.ToBytes(),
		HeaderPrv:    hdrPrv,
		PayloadPrv:   pldPrv,
		Deadline:     time.Now().Add(deadline),
	}
	return
}

// AddBytes adds to the Remaining counter, used when new data allowance has been
// purchased.
func (s *Session) AddBytes(b uint64) { s.Remaining += b }

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

type Sessions []*Session

func (s Sessions) Len() int { return len(s) }

func (s Sessions) Add(se *Session) Sessions { return append(s, se) }

func (s Sessions) Delete(se *Session) Sessions {
	for i := range s {
		if s[i] == se {
			return append(s[:i], s[i:]...)
		}
	}
	return s
}

func (s Sessions) DeleteByID(id nonce.ID) Sessions {
	for i := range s {
		if s[i].ID == id {
			return append(s[:i], s[i:]...)
		}
	}
	return s
}

func (s Sessions) Find(t nonce.ID) (se *Session) {
	for i := range s {
		if s[i].ID == t {
			se = s[i]
			return
		}
	}
	return
}

func (s Sessions) FindPub(pubKey *pub.Key) (se *Session) {
	for i := range s {
		if s[i].HeaderPub.Equals(pubKey) {
			se = s[i]
			return
		}
	}
	return
}
