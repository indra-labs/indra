package engine

import (
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

type Codec interface {
	Magic() string
	Encode(s *Splice) (e error)
	Decode(s *Splice) (e error)
	Len() int
	GetMung() Mung
}

// Mung is an interface messages that can be layered over each other and have
// a set of processing instructions for the data in them, and, if relevant,
// how to account for them in sessions.
type Mung interface {
	Codec
	Wrap(inner Mung)
	Handle(s *Splice, p Mung, ng *Engine) (e error)
	Account(res *SendData, sm *SessionManager, s *SessionData,
		last bool) (skip bool, sd *SessionData)
}

type Transport interface {
	Send(b slice.Bytes) (e error)
	Receive() <-chan slice.Bytes
}
