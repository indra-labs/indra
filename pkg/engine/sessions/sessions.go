// Package sessions defines some key data structures relating to the data for sessions, imported by sess package for reading and writing session and circuit metadata.
package sessions

import (
	"fmt"
	"git.indra-labs.org/dev/ind/pkg/crypto"
	"git.indra-labs.org/dev/ind/pkg/crypto/nonce"
	"git.indra-labs.org/dev/ind/pkg/crypto/sha256"
	"git.indra-labs.org/dev/ind/pkg/engine/node"
	log2 "git.indra-labs.org/dev/ind/pkg/proc/log"
	"github.com/gookit/color"
	"github.com/lightningnetwork/lnd/lnwire"
)

var (
	log   = log2.GetLogger()
	fails = log.E.Chk
)

type (
	// Sessions are arbitrary length lists of Data.
	Sessions []*Data

	// A Circuit is the generic fixed-length path used for most messages.
	Circuit [5]*Data

	// A Data keeps track of a connection session. It specifically maintains
	// the account of available bandwidth allocation before it needs to be recharged
	// with new credit, and the current state of the encryption.
	Data struct {

		// ID is the internal reference used by other data structures referring to this
		// one.
		ID nonce.ID

		// Node is the node.Node that is providing the relaying service.
		Node *node.Node

		// Remaining is the current balance on the session.
		Remaining lnwire.MilliSatoshi

		// Header and Payload are the two key sets used for data relayed for this
		// session. Header keys are embedded in the 3 layer RoutingHeader, and Payload
		// keys are only used to derive the secrets used with the given public sender key
		// that enables the Exit and Hidden Service to encrypt replies that then get
		// unwrapped successively on the return path.
		Header, Payload *crypto.Keys

		// Preimage is essentially the hash of the bytes of the Header and Payload keys.
		// This enables the relay to associate a payment with a session key pair and thus
		// to be able to account the client's usage.
		Preimage sha256.Hash

		// Hop is the position at which this session is used, private and secret to the
		// client. Sessions are prescribed to be used at a given position only, in order
		// to prevent any cross correlations being visible.
		Hop byte
	}
)

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
		color.Yellow.Sprint(s.Node.Addresses[0].String()), who,
		typ, s.Header.Bytes.String(), s.Remaining, sats)
	s.Remaining -= sats
	return true
}

// IncSats adds to the Remaining counter, used when new data allowance has been
// purchased.
func (s *Data) IncSats(sats lnwire.MilliSatoshi, sender bool, typ string) {
	who := "relay"
	if sender {
		who = "client"
	}
	log.D.F("%s session %d %x current %v incrementing by %v", who, typ,
		s.Header.Bytes.String(), s.Remaining, sats)
	s.Remaining += sats
}

func (s *Data) String() string {
	return fmt.Sprintf("%s sesssion %s node %s hop %d",
		s.Node.Addresses[0].String(), s.Header.Bytes.String(),
		s.Node.ID, s.Hop)
}

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

// String is a stringer for Circuits that attempts to make them readable as like
// a table.
func (c Circuit) String() (o string) {
	o += "[ "
	for i := range c {
		if c[i] == nil {
			o += "              "
		} else {
			o += c[i].Header.Bytes.String() + " "
		}
	}
	o += "]"
	return
}
