package sessions

import (
	"fmt"
	
	"git-indra.lan/indra-labs/lnd/lnd/lnwire"
	"github.com/gookit/color"
	
	"git-indra.lan/indra-labs/indra"
	"git-indra.lan/indra-labs/indra/pkg/crypto"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/engine/node"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	fails = log.E.Chk
)

// A Data keeps track of a connection session. It specifically maintains
// the account of available bandwidth allocation before it needs to be recharged
// with new credit, and the current state of the encryption.
type Data struct {
	ID              nonce.ID
	Node            *node.Node
	Remaining       lnwire.MilliSatoshi
	Header, Payload *crypto.Keys
	Preimage        sha256.Hash
	Hop             byte
}

func (s *Data) String() string {
	return fmt.Sprintf("%s sesssion %s node %s hop %d",
		s.Node.AddrPort.String(), s.ID,
		s.Node.ID, s.Hop)
}

// A Circuit is the generic fixed-length path used for most messages.
type Circuit [5]*Data

func (c Circuit) String() (o string) {
	o += "[ "
	for i := range c {
		if c[i] == nil {
			o += "              "
		} else {
			o += c[i].ID.String() + " "
		}
	}
	o += "]"
	return
}

// Sessions are arbitrary length lists of Data.
type Sessions []*Data

// NewSessionData creates a new Data, generating cached public key bytes
// and preimage.
func NewSessionData(
	id nonce.ID,
	node *node.Node,
	rem lnwire.MilliSatoshi,
	hdr, pld *crypto.Keys,
	hop byte,
) (s *Data) {
	
	var e error
	if hdr == nil || pld == nil {
		if hdr, pld, e = crypto.Generate2Keys(); fails(e) {
		}
	}
	h, p := hdr.Prv.ToBytes(), pld.Prv.ToBytes()
	s = &Data{
		ID:        id,
		Node:      node,
		Remaining: rem,
		Header:    hdr,
		Payload:   pld,
		Preimage:  sha256.Single(append(h[:], p[:]...)),
		Hop:       hop,
	}
	return
}

// IncSats adds to the Remaining counter, used when new data allowance has been
// purchased.
func (s *Data) IncSats(sats lnwire.MilliSatoshi, sender bool, typ string) {
	who := "relay"
	if sender {
		who = "client"
	}
	log.D.F("%s session %d %x current %v incrementing by %v", who, typ, s.ID,
		s.Remaining, sats)
	s.Remaining += sats
}

// DecSats reduces the amount Remaining, if the requested amount would put the
// total below zero it returns false, signalling that new data allowance needs
// to be purchased before any further messages can be sent.
func (s *Data) DecSats(sats lnwire.MilliSatoshi, sender bool,
	typ string) bool {
	
	if s.Remaining < sats {
		return false
	}
	who := "relay"
	if sender {
		who = "client"
	}
	log.D.F("%s %s session %s %s current %v decrementing by %v",
		color.Yellow.Sprint(s.Node.AddrPort.String()), who,
		typ, s.ID,
		s.Remaining, sats)
	s.Remaining -= sats
	return true
}
