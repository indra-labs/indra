package engine

import (
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/engine/types"
	"git-indra.lan/indra-labs/indra/pkg/util/octet"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

const (
	ResponseMagic = "rs"
	ResponseLen   = MagicLen + slice.Uint32Len + slice.Uint16Len +
		nonce.IDLen + 1
)

type Response struct {
	nonce.ID
	Port uint16
	Load byte
	slice.Bytes
}

var responsePrototype types.Onion = &Response{}

func init() { Register(ResponseMagic, responsePrototype) }

func (o Skins) Response(id nonce.ID, res slice.Bytes, port uint16) Skins {
	rs := Response{ID: id, Port: port, Bytes: res}
	return append(o, &rs)
}

func (x *Response) Magic() string { return ResponseMagic }

func (x *Response) Encode(s *octet.Splice) (e error) {
	return s.
		Magic(ResponseMagic).
		ID(x.ID).
		Uint16(x.Port).
		Byte(x.Load).
		Bytes(x.Bytes)
}

func (x *Response) Decode(s *octet.Splice) (e error) {
	if e = TooShort(s.Remaining(), ResponseLen-MagicLen,
		ResponseMagic); check(e) {
		return
	}
	return s.
		ReadID(&x.ID).
		ReadUint16(&x.Port).
		ReadByte(&x.Load).
		ReadBytes(&x.Bytes)
}

func (x *Response) Len() int { return ResponseLen + len(x.Bytes) }

func (x *Response) Wrap(inner types.Onion) {}

func (x *Response) Handle(s *octet.Splice, p types.Onion,
	ng *Engine) (e error) {
	
	pending := ng.PendingResponses.Find(x.ID)
	log.T.F("searching for pending ID %x", x.ID)
	if pending != nil {
		for i := range pending.Billable {
			se := ng.FindSession(pending.Billable[i])
			if se != nil {
				typ := "response"
				relayRate := se.RelayRate
				dataSize := s.Len()
				switch i {
				case 0, 1:
					dataSize = pending.SentSize
				case 2:
					for j := range se.Services {
						if se.Services[j].Port == x.Port {
							relayRate = se.Services[j].RelayRate / 2
							typ = "exit"
						}
					}
				}
				ng.DecSession(se.ID, relayRate*dataSize, true, typ)
			}
		}
		ng.PendingResponses.ProcessAndDelete(x.ID, x.Bytes)
	}
	return
}
