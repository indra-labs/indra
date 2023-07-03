// Package responses handles waiting for and responding to received responses, including tracking the session billing and custom callback hooks when responses arrive.
package responses

import (
	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/engine/sessions"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/util/qu"
	"github.com/indra-labs/indra/pkg/util/slice"
	"sync"
	"time"
)

var (
	log   = log2.GetLogger()
	fails = log.E.Chk
)

type (
	// Response is a record stored in anticipation of a response coming back from the
	// peer processing an Exit or message for a service available at a relay.
	Response struct {

		// ID is an internal reference code for associating data together.
		ID nonce.ID

		// SentSize is the number of bytes total in the outgoing message (before
		// padding).
		SentSize int

		// Port is the destination port of the service the Exit or other reply-enabling
		// onion is intended for.
		Port uint16

		// Billable is a list of the public keys of the sessions that are used in the
		// circuit.
		Billable []crypto.PubBytes

		// Return is the last hop return session used for this message.
		Return crypto.PubBytes

		// PostAcct is a slice of functions that action the balance adjustment of the
		// message for each of the hops in the circuit. This enables the engine to
		// contact the relays in a circuit in outbound order to both correctly update the
		// session balance and diagnose which peer the message did not get to (due to
		// being offline or congested).
		PostAcct []func()

		// Sessions that were used in the circuit.
		sessions.Sessions

		// Callback is a hook added to the dispatch for the message that will be executed with the reassembled response packet.
		Callback Callback

		// Time records the time this message was dispatched.
		Time time.Time

		// Success channel is closed when the transmission is successful. todo: nothing is actually using this?
		Success qu.C
	}

	// ResponseParams is the inputs required to construct a Response.
	ResponseParams struct {
		ID       nonce.ID
		SentSize int
		S        sessions.Sessions
		Billable []crypto.PubBytes
		Ret      crypto.PubBytes
		Port     uint16
		Callback Callback
		PostAcct []func()
	}

	// Callback is a function signature of the callback hook to be added to the
	// dispatch parameters that is intended to be called when the response arrives.
	// This is distinct from the billing callbacks, this is application-specific.
	Callback func(id nonce.ID, ifc interface{}, b slice.Bytes) (e error)

	// Pending is a mutex protected data structure for storing pending Response data.
	Pending struct {

		// This data must not be read and written concurrently.
		sync.Mutex

		// The collection of responses still waiting.
		responses []*Response
	}
)

// Add a new response using the ResponseParams.
func (p *Pending) Add(pr ResponseParams) {
	p.Lock()
	defer p.Unlock()
	log.T.F("adding response hook %s", pr.ID)
	r := &Response{
		ID:       pr.ID,
		SentSize: pr.SentSize,
		Time:     time.Now(),
		Billable: pr.Billable,
		Return:   pr.Ret,
		Port:     pr.Port,
		PostAcct: pr.PostAcct,
		Callback: pr.Callback,
		Success:  qu.T(),
	}
	p.responses = append(p.responses, r)
}

// Find the response with the matching nonce.ID.
func (p *Pending) Find(id nonce.ID) (pr *Response) {
	//p.Lock()
	//defer p.Unlock()
	for i := range p.responses {
		if p.responses[i].ID == id {
			return p.responses[i]
		}
	}
	return
}

// FindOlder returns all Response data that is is older than a specified time.
func (p *Pending) FindOlder(t time.Time) (r []*Response) {
	p.Lock()
	defer p.Unlock()
	for i := range p.responses {
		if p.responses[i].Time.Before(t) {
			r = append(r, p.responses[i])
		}
	}
	return
}

// GetOldestPending returns the oldest pending response. In fact, this is just
// the last one with the lowest index, because they are not sorted and are
// always added in chronological order.
//
// New items are always added to the end, and so the first index is the oldest in
// the slice.
func (p *Pending) GetOldestPending() (pr *Response) {
	p.Lock()
	defer p.Unlock()
	if len(p.responses) > 0 {
		// Pending responses are added in chronological order to the end so the
		// first one in the slice is the oldest.
		return p.responses[0]
	}
	return
}

// ProcessAndDelete runs the callback and post accounting function list and
// deletes the pending response.
//
// Returns true if it found and deleted a pending response.
func (p *Pending) ProcessAndDelete(id nonce.ID, ifc interface{},
	b slice.Bytes) (found bool, e error) {
	p.Lock()
	defer p.Unlock()
	for i := range p.responses {
		if p.responses[i].ID == id {
			log.D.F("deleting response %s", id)
			// Stop the timeout handler.
			p.responses[i].Success.Q()
			for _, fn := range p.responses[i].PostAcct {
				fn()
			}
			e = p.responses[i].Callback(id, ifc, b)
			if i < 1 {
				p.responses = p.responses[1:]
			} else {
				p.responses = append(p.responses[:i],
					p.responses[i+1:]...)
			}
			found = true
			break
		}
	}
	return
}
