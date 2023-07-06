// Package exit is an onion message type that contains a payload intended to be forwarded to the exit service of a relay.
package exit

import (
	"github.com/indra-labs/indra/pkg/codec"
	"github.com/indra-labs/indra/pkg/codec/onion/crypt"
	"github.com/indra-labs/indra/pkg/codec/onion/end"
	"github.com/indra-labs/indra/pkg/codec/onion/response"
	"github.com/indra-labs/indra/pkg/codec/ont"
	"github.com/indra-labs/indra/pkg/codec/reg"
	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/crypto/sha256"
	"github.com/indra-labs/indra/pkg/engine/magic"
	"github.com/indra-labs/indra/pkg/engine/sess"
	"github.com/indra-labs/indra/pkg/engine/sessions"
	"github.com/indra-labs/indra/pkg/hidden"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/util/slice"
	"github.com/indra-labs/indra/pkg/util/splice"
	"reflect"
	"time"
)

var (
	log   = log2.GetLogger()
	fails = log.E.Chk
)

const (
	Magic = "exit"
	Len   = magic.Len +
		slice.Uint16Len +
		3*sha256.Len +
		slice.Uint32Len +
		nonce.IVLen*3 +
		nonce.IDLen
)

// Exit is a
type Exit struct {

	// ID is the identifier that will be embedded with the response to this message
	// relayed from the exit service.
	ID nonce.ID

	// Ciphers is a set of 3 symmetric ciphers that are to be used in their
	// given order over the reply message from the service.
	crypto.Ciphers

	// Nonces are the nonces to use with the cipher when creating the
	// encryption for the reply message,
	// they are common with the crypts in the header.
	crypto.Nonces

	// Port identifies the type of service as well as being the port used by the
	// service to be relayed to. This should be the well-known protocol associated
	// with the port number, eg 80 for HTTP, 53 for DNS, etc.
	Port uint16

	// Bytes are the message to be passed to the exit service.
	slice.Bytes

	// Onion contains the rest of the message, in this case the reply RoutingHeader.
	ont.Onion
}

// Account searches for the relevant service, applies the balance change to the
// account that will be in effect when the response has arrived and delivery is
// confirmed.
func (x *Exit) Account(res *sess.Data, sm *sess.Manager,
	s *sessions.Data, last bool) (skip bool, sd *sessions.Data) {
	s.Node.Lock()
	for j := range s.Node.Services {
		if s.Node.Services[j].Port != x.Port {
			s.Node.Unlock()
			continue
		}
		s.Node.Unlock()
		res.Port = x.Port
		res.PostAcct = append(res.PostAcct,
			func() {
				sm.DecSession(s.Header.Bytes,
					int(s.Node.Services[j].RelayRate)*len(res.B)/2, true, "exit")
			})
		break
	}
	res.Billable = append(res.Billable, s.Header.Bytes)
	res.ID = x.ID
	skip = true
	return
}

// Decode what should be an Exit message from a splice.Splice.
func (x *Exit) Decode(s *splice.Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), Len-magic.Len,
		Magic); fails(e) {
		return
	}
	s.ReadID(&x.ID).
		ReadCiphers(&x.Ciphers).
		ReadNonces(&x.Nonces).
		ReadUint16(&x.Port).
		ReadBytes(&x.Bytes)
	return
}

// Encode this Exit into a splice.Splice's next bytes.
func (x *Exit) Encode(s *splice.Splice) (e error) {
	log.T.S("encoding", reflect.TypeOf(x),
		x.ID, x.Ciphers, x.Nonces, x.Port, x.Bytes.ToBytes(),
	)
	return x.Onion.Encode(s.
		Magic(Magic).
		ID(x.ID).Ciphers(x.Ciphers).Nonces(x.Nonces).
		Uint16(x.Port).
		Bytes(x.Bytes),
	)
}

// Unwrap returns the onion inside this Exit message.
func (x *Exit) Unwrap() interface{} { return x.Onion }

// Handle provides the relay switching logic for an engine handling an Exit
// message.
func (x *Exit) Handle(s *splice.Splice, p ont.Onion, ng ont.Ngin) (e error) {
	// payload is forwarded to a local port and the result is forwarded
	// back with a reverse header.
	var result slice.Bytes
	h := sha256.Single(x.Bytes)
	log.T.S(h)
	log.T.F("%s received exit id %s", ng.Mgr().
		GetLocalNodeAddressString(), x.ID)
	if e = ng.Mgr().SendFromLocalNode(x.Port, x.Bytes); fails(e) {
		return
	}
	timer := time.NewTicker(time.Second * 5)
	select {
	case result = <-ng.Mgr().GetLocalNode().ReceiveFrom(x.Port):
	case <-timer.C:
	}
	// We need to wrap the result in a message crypt.
	res := ont.Encode(&response.Response{
		ID:    x.ID,
		Port:  x.Port,
		Load:  byte(ng.GetLoad()),
		Bytes: result,
	})
	rb := hidden.FormatReply(hidden.GetRoutingHeaderFromCursor(s),
		x.Ciphers, x.Nonces, res.GetAll())
	switch on := p.(type) {
	case *crypt.Crypt:
		sess := ng.Mgr().FindSessionByHeader(on.ToPriv)
		if sess == nil {
			return
		}
		sess.Node.Lock()
		for i := range sess.Node.Services {
			if x.Port != sess.Node.Services[i].Port {
				sess.Node.Unlock()
				continue
			}
			in := int(sess.Node.Services[i].RelayRate) * s.Len() / 2
			out := int(sess.Node.Services[i].RelayRate) * rb.Len() / 2
			sess.Node.Unlock()
			ng.Mgr().DecSession(sess.Header.Bytes, in+out, false, "exit")
			break
		}
	}
	ng.HandleMessage(rb, x)
	return
}

// Len returns the length of this Exit message (payload and return header Onion
// included.
func (x *Exit) Len() int { return Len + x.Bytes.Len() + x.Onion.Len() }

// Magic is the identifying 4 byte string indicating an Exit message follows.
func (x *Exit) Magic() string { return Magic }

// Params are the parameters to generate an Exit onion.
type Params struct {
	Port       uint16
	Payload    slice.Bytes
	ID         nonce.ID
	Alice, Bob *sessions.Data
	S          sessions.Circuit
	KS         *crypto.KeySet
}

// ExitPoint is the return routing parameters delivered inside an Exit onion.
type ExitPoint struct {

	// Routing contains all the information required to generate a RoutingHeader and
	// cipher/nonce set.
	*Routing

	// ReturnPubs are the public keys of the session payload, which are not in the
	// message but previously shared to create a session.
	ReturnPubs crypto.Pubs
}

// Wrap puts another onion inside this Exit onion.
func (x *Exit) Wrap(inner ont.Onion) { x.Onion = inner }

// Routing is the sessions and keys required to generate a return RoutingHeader.
type Routing struct {

	// Sessions that the RoutingHeader is using.
	Sessions [3]*sessions.Data

	// Keys being used to form the other ECDH half for the encryption.
	Keys crypto.Privs

	// The three nonces that match up with each of the three session RoutingHeader
	// Crypt nonces.
	crypto.Nonces
}

// New creates a new Exit onion.
func New(id nonce.ID, port uint16, payload slice.Bytes,
	ep *ExitPoint) ont.Onion {
	return &Exit{
		ID:      id,
		Ciphers: crypto.GenCiphers(ep.Keys, ep.ReturnPubs),
		Nonces:  ep.Nonces,
		Port:    port,
		Bytes:   payload,
		Onion:   &end.End{},
	}
}

// Gen is a factory function for an Exit.
func Gen() codec.Codec { return &Exit{} }

func init() { reg.Register(Magic, Gen) }
