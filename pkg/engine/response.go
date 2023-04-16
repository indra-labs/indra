package engine

import (
	"reflect"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/engine/coding"
	"git-indra.lan/indra-labs/indra/pkg/engine/magic"
	"git-indra.lan/indra-labs/indra/pkg/splice"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

const (
	ResponseMagic = "rs"
	ResponseLen   = magic.Len + slice.Uint32Len + slice.Uint16Len +
		nonce.IDLen + 1
)

type Response struct {
	ID   nonce.ID
	Port uint16
	Load byte
	slice.Bytes
}

func responseGen() coding.Codec           { return &Response{} }
func init()                               { Register(ResponseMagic, responseGen) }
func (x *Response) Magic() string         { return ResponseMagic }
func (x *Response) Len() int              { return ResponseLen + len(x.Bytes) }
func (x *Response) Wrap(inner Onion)      {}
func (x *Response) GetOnion() interface{} { return x }

func (x *Response) Encode(s *splice.Splice) (e error) {
	log.T.S("encoding", reflect.TypeOf(x),
		x.ID, x.Port, x.Load, x.Bytes.ToBytes(),
	)
	s.
		Magic(ResponseMagic).
		ID(x.ID).
		Uint16(x.Port).
		Byte(x.Load).
		Bytes(x.Bytes)
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

func (x *Response) Handle(s *splice.Splice, p Onion,
	ni interface{}) (e error) {
	
	ng := ni.(*Engine)
	pending := ng.PendingResponses.Find(x.ID)
	log.T.F("searching for pending ID %s", x.ID)
	if pending != nil {
		for i := range pending.Billable {
			se := ng.FindSession(pending.Billable[i])
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
				ng.DecSession(se.ID, relayRate*dataSize, true, typ)
			}
		}
		ng.PendingResponses.ProcessAndDelete(x.ID, nil, x.Bytes)
	}
	return
}

func (x *Response) Account(res *Data, sm *SessionManager,
	s *SessionData, last bool) (skip bool, sd *SessionData) {
	return
}
