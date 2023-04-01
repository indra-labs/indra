package engine

import (
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

type Codec interface {
	Encode(s *Splice) (e error)
	Decode(s *Splice) (e error)
	Len() int
}

// Onion is an interface for the layers of messages each encrypted inside a
// OnionSkin, which provides the cipher for the inner layers inside it.
type Onion interface {
	Magic() string
	Codec
	Wrap(inner Onion)
	Handle(s *Splice, p Onion, ng *Engine) (e error)
	Account(res *SendData, sm *SessionManager, s *SessionData,
		last bool) (skip bool, sd *SessionData)
}

type Transport interface {
	Send(b slice.Bytes) (e error)
	Receive() <-chan slice.Bytes
}
