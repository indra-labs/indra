package node

import (
	"github.com/indra-labs/indra/pkg/key/prv"
	"github.com/indra-labs/indra/pkg/key/pub"
	"github.com/indra-labs/indra/pkg/lnwire"
	"github.com/indra-labs/indra/pkg/nonce"
	"github.com/indra-labs/indra/pkg/sha256"
)

// A Session keeps track of a connection session. It specifically maintains the
// account of available bandwidth allocation before it needs to be recharged
// with new credit, and the current state of the encryption.
type Session struct {
	nonce.ID
	*Node
	Remaining                 lnwire.MilliSatoshi
	HeaderPrv, PayloadPrv     *prv.Key
	HeaderPub, PayloadPub     *pub.Key
	HeaderBytes, PayloadBytes pub.Bytes
	Preimage                  sha256.Hash
	Hop                       byte
}

// NewSession creates a new Session.
//
// Purchasing a session the seller returns a token, based on a requested data
// allocation.
func NewSession(
	id nonce.ID,
	node *Node,
	rem lnwire.MilliSatoshi,
	hdrPrv *prv.Key,
	pldPrv *prv.Key,
	hop byte,
) (s *Session) {

	var e error
	if hdrPrv == nil || pldPrv == nil {
		if hdrPrv, e = prv.GenerateKey(); check(e) {
		}
		if pldPrv, e = prv.GenerateKey(); check(e) {
		}
	}
	hdrPub := pub.Derive(hdrPrv)
	pldPub := pub.Derive(pldPrv)
	h, p := hdrPrv.ToBytes(), pldPrv.ToBytes()
	s = &Session{
		ID:           id,
		Node:         node,
		Remaining:    rem,
		HeaderPub:    hdrPub,
		HeaderBytes:  hdrPub.ToBytes(),
		PayloadPub:   pldPub,
		PayloadBytes: pldPub.ToBytes(),
		HeaderPrv:    hdrPrv,
		PayloadPrv:   pldPrv,
		Preimage:     sha256.Single(append(h[:], p[:]...)),
		Hop:          hop,
	}
	return
}

// AddBytes adds to the Remaining counter, used when new data allowance has been
// purchased.
func (s *Session) AddBytes(b lnwire.MilliSatoshi) { s.Remaining += b }

// SubtractBytes reduces the amount Remaining, if the requested amount would put
// the total below zero it returns false, signalling that new data allowance
// needs to be purchased before any further messages can be sent.
func (s *Session) SubtractBytes(b lnwire.MilliSatoshi) bool {
	if s.Remaining < b {
		return false
	}
	s.Remaining -= b
	return true
}

type Circuit [5]*Session

type Sessions []*Session
