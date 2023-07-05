// Package headers is a set of helpers for defining the data to put in a reverse message to enable source routed messages to return anonymously to the client who sent them. These are used for exits and hidden service routing headers.
package headers

import (
	"github.com/indra-labs/indra/pkg/codec/onion/exit"
	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/engine/sessions"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
)

var (
	log   = log2.GetLogger()
	fails = log.E.Chk
)

// Headers is a collection of keys and sessions required to construct reply
// headers for the return path, including the forward and return paths.
type Headers struct {

	// Forward and Return - the sessions in the forward hops and return hops.
	Forward, Return *exit.Routing

	// ReturnPubs
	ReturnPubs crypto.Pubs
}

// GetHeaders returns a Headers constructed using a (partially preloaded)
// circuit, the client's node, and the node of the exit, and all the session keys
// required for each layer of encryption.
func GetHeaders(alice, bob *sessions.Data, c sessions.Circuit,
	ks *crypto.KeySet) (h *Headers) {

	fwKeys := ks.Next3()
	rtKeys := ks.Next3()
	n := crypto.GenNonces(6)
	var rtNonces, fwNonces [3]nonce.IV
	copy(fwNonces[:], n[:3])
	copy(rtNonces[:], n[3:])
	var fwSessions, rtSessions [3]*sessions.Data
	copy(fwSessions[:], c[:2])
	fwSessions[2] = bob
	copy(rtSessions[:], c[3:])
	rtSessions[2] = alice
	var returnPubs crypto.Pubs
	returnPubs[0] = c[3].Payload.Pub
	returnPubs[1] = c[4].Payload.Pub
	returnPubs[2] = alice.Payload.Pub
	h = &Headers{
		Forward: &exit.Routing{
			Sessions: fwSessions,
			Keys:     fwKeys,
			Nonces:   fwNonces,
		},
		Return: &exit.Routing{
			Sessions: rtSessions,
			Keys:     rtKeys,
			Nonces:   rtNonces,
		},
		ReturnPubs: returnPubs,
	}
	return
}

// ExitPoint is similar to GetHeaders except it does not include the forward
// path.
func (h *Headers) ExitPoint() *exit.ExitPoint {
	return &exit.ExitPoint{
		Routing:    h.Return,
		ReturnPubs: h.ReturnPubs,
	}
}
