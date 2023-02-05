package traffic

import (
	"encoding/hex"

	"git-indra.lan/indra-labs/lnd/lnd/lnwire"

	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
)

// A Session keeps track of a connection session. It specifically maintains the
// account of available bandwidth allocation before it needs to be recharged
// with new credit, and the current state of the encryption.
type Session struct {
	ID nonce.ID
	*Node
	Remaining                 lnwire.MilliSatoshi
	HeaderPrv, PayloadPrv     *prv.Key
	HeaderPub, PayloadPub     *pub.Key
	HeaderBytes, PayloadBytes pub.Bytes
	Preimage                  sha256.Hash
	Hop                       byte
}

// NewSession creates a new Session, generating cached public key bytes and
// preimage.
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

// IncSats adds to the Remaining counter, used when new data allowance has been
// purchased.
func (s *Session) IncSats(sats lnwire.MilliSatoshi,
	sender bool, typ string) {
	who := "relay"
	if sender {
		who = "client"
	}
	log.D.F("%s session %d %x current %v incrementing by %v", who, typ, s.ID,
		s.Remaining, sats)
	s.Remaining += sats
}

// DecSats reduces the amount Remaining, if the requested amount would put
// the total below zero it returns false, signalling that new data allowance
// needs to be purchased before any further messages can be sent.
func (s *Session) DecSats(sats lnwire.MilliSatoshi,
	sender bool, typ string) bool {
	if s.Remaining < sats {
		return false
	}
	who := "relay"
	if sender {
		who = "client"
	}
	log.D.F("%s session %s %x current %v decrementing by %v", who, typ, s.ID,
		s.Remaining, sats)
	s.Remaining -= sats
	return true
}

// A Circuit is the generic fixed length path used for most messages.
type Circuit [5]*Session

func (c Circuit) String() (o string) {
	o += "[ "
	for i := range c {
		if c[i] == nil {
			o += "                 "
		} else {
			o += hex.EncodeToString(c[i].ID[:]) + " "
		}
	}
	o += "]"
	return
}

// Sessions are arbitrary length lists of sessions.
type Sessions []*Session
