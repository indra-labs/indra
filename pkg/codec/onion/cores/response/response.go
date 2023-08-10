// Package response provides a message type in response to an Exit message.
package response

import (
	"git.indra-labs.org/dev/ind/pkg/codec"
	"git.indra-labs.org/dev/ind/pkg/codec/ont"
	"git.indra-labs.org/dev/ind/pkg/codec/reg"
	"git.indra-labs.org/dev/ind/pkg/crypto/nonce"
	"git.indra-labs.org/dev/ind/pkg/engine/magic"
	"git.indra-labs.org/dev/ind/pkg/engine/sess"
	"git.indra-labs.org/dev/ind/pkg/engine/sessions"
	log2 "git.indra-labs.org/dev/ind/pkg/proc/log"
	"git.indra-labs.org/dev/ind/pkg/util/slice"
	"git.indra-labs.org/dev/ind/pkg/util/splice"
	"reflect"
)

var (
	log   = log2.GetLogger()
	fails = log.E.Chk
)

const (
	Magic = "resp"
	Len   = magic.Len +
		slice.Uint32Len +
		slice.Uint16Len +
		nonce.IDLen + 1
)

// Response is a reply to an Exit message.
type Response struct {

	// ID of the Exit message.
	ID nonce.ID

	// Port of the Exit service.
	Port uint16

	// Load of the Exit relay.
	Load byte

	// Bytes of the Response to the Exit request message.
	slice.Bytes
}

func New(id nonce.ID, port uint16, res slice.Bytes, load byte) ont.Onion {
	return &Response{ID: id, Port: port, Bytes: res, Load: load}
}

// Account for an Response.
//
// TODO: this is supposed to affect half the exit fee and the two hops following.
func (x *Response) Account(res *sess.Data, sm *sess.Manager,
	s *sessions.Data, last bool) (skip bool, sd *sessions.Data) {
	return
}

// Decode what should be an Response message from a splice.Splice.
func (x *Response) Decode(s *splice.Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), Len-magic.Len,
		Magic); fails(e) {
		return
	}
	s.
		ReadID(&x.ID).
		ReadUint16(&x.Port).
		ReadByte(&x.Load).
		ReadBytes(&x.Bytes)
	return
}

// Encode this Response into a splice.Splice's next bytes.
func (x *Response) Encode(s *splice.Splice) (e error) {
	log.T.Ln("encoding", reflect.TypeOf(x)) // x.Keys, x.Port, x.Load, x.Bytes.ToBytes(),

	s.
		Magic(Magic).
		ID(x.ID).
		Uint16(x.Port).
		Byte(x.Load).
		Bytes(x.Bytes)
	return
}

// Unwrap is a no-op because there is no onion inside a Response (that we
// concern ourselves with).
func (x *Response) Unwrap() interface{} { return x }

// Handle provides the relay switching logic for an engine handling an Response
// message.
func (x *Response) Handle(s *splice.Splice, p ont.Onion, ng ont.Ngin) (e error) {
	pending := ng.Pending().Find(x.ID)
	log.T.F("searching for pending Keys %s", x.ID)
	if pending != nil {
		for i := range pending.Billable {
			se := ng.Mgr().FindSessionByPubkey(pending.Billable[i])
			if se != nil {
				typ := "response"
				relayRate := se.Node.RelayRate
				dataSize := s.Len()
				switch i {
				case 0, 1:
					dataSize = pending.SentSize
				case 2:
					se.Node.Lock()
					for j := range se.Node.Services {
						if se.Node.Services[j].Port == x.Port {
							relayRate = se.Node.Services[j].RelayRate / 2
							typ = "exit"
							break
						}
					}
					se.Node.Unlock()
				}
				ng.Mgr().DecSession(se.Header.Bytes, int(relayRate)*dataSize, true, typ)
			}
		}
		ng.Pending().ProcessAndDelete(x.ID, nil, x.Bytes)
	}
	return
}

// Len returns the length of the onion starting from this one (used to size a
// Splice).
func (x *Response) Len() int {

	codec.MustNotBeNil(x)

	return Len + len(x.Bytes)
}

// Magic bytes identifying a HiddenService message is up next.
func (x *Response) Magic() string { return Magic }

// Wrap places another onion inside this one in its slot. Which isn't going to
// happen.
func (x *Response) Wrap(inner ont.Onion) {}

// Gen is a factory function for an IntroQuery.
func Gen() codec.Codec { return &Response{} }

func init() { reg.Register(Magic, Gen) }
