package engine

import (
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

// Codec is a unit of data that can be read and written from a binary form. All
// Mung are Codec but not all Codec are Mung. Codec is also used for the
// Dispatcher's message headers.
type Codec interface {
	Magic() string
	Encode(s *Splice) (e error)
	Decode(s *Splice) (e error)
	Len() int
	GetMung() Mung
}

// Mung are messages that can be layered over each other and have
// a set of processing instructions for the data in them, and, if relevant,
// how to account for them in sessions.
type Mung interface {
	Codec
	Wrap(inner Mung)
	Handle(s *Splice, p Mung, ng *Engine) (e error)
	Account(res *SendData, sm *SessionManager, s *SessionData,
		last bool) (skip bool, sd *SessionData)
}

// Transport is a generic interface for sending and receiving slices of bytes.
type Transport interface {
	Send(b slice.Bytes) (e error)
	Receive() <-chan slice.Bytes
}
