package onions

import (
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/engine/coding"
	"github.com/indra-labs/indra/pkg/engine/magic"
	"github.com/indra-labs/indra/pkg/engine/sess"
	"github.com/indra-labs/indra/pkg/engine/sessions"
	"github.com/indra-labs/indra/pkg/onions/reg"
	"github.com/indra-labs/indra/pkg/util/slice"
	"github.com/indra-labs/indra/pkg/util/splice"
	"reflect"
)

const (
	ResponseMagic = "resp"
	ResponseLen   = magic.Len +
		slice.Uint32Len +
		slice.Uint16Len +
		nonce.IDLen + 1
)

type Response struct {
	ID   nonce.ID
	Port uint16
	Load byte
	slice.Bytes
}

func NewResponse(id nonce.ID, port uint16, res slice.Bytes, load byte) Onion {
	return &Response{ID: id, Port: port, Bytes: res, Load: load}
}

func (x *Response) Account(res *sess.Data, sm *sess.Manager,
	s *sessions.Data, last bool) (skip bool, sd *sessions.Data) {
	return
}

func (x *Response) Decode(s *splice.Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), ResponseLen-magic.Len,
		ResponseMagic); fails(e) {
		return
	}
	s.
		ReadID(&x.ID).
		ReadUint16(&x.Port).
		ReadByte(&x.Load).
		ReadBytes(&x.Bytes)
	return
}

func (x *Response) Encode(s *splice.Splice) (e error) {
	log.T.Ln("encoding", reflect.TypeOf(x)) // x.Keys, x.Port, x.Load, x.Bytes.ToBytes(),

	s.
		Magic(ResponseMagic).
		ID(x.ID).
		Uint16(x.Port).
		Byte(x.Load).
		Bytes(x.Bytes)
	return
}

func (x *Response) GetOnion() interface{} { return x }

func (x *Response) Handle(s *splice.Splice, p Onion, ng Ngin) (e error) {
	pending := ng.Pending().Find(x.ID)
	log.T.F("searching for pending Keys %s", x.ID)
	if pending != nil {
		for i := range pending.Billable {
			se := ng.Mgr().FindSession(pending.Billable[i])
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
				ng.Mgr().DecSession(se.Header.Bytes, relayRate*dataSize, true, typ)
			}
		}
		ng.Pending().ProcessAndDelete(x.ID, nil, x.Bytes)
	}
	return
}

func (x *Response) Len() int         { return ResponseLen + len(x.Bytes) }
func (x *Response) Magic() string    { return ResponseMagic }
func (x *Response) Wrap(inner Onion) {}
func init()                     { reg.Register(ResponseMagic, responseGen) }
func responseGen() coding.Codec { return &Response{} }
