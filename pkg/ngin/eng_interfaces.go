package ngin

import (
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
	"git-indra.lan/indra-labs/indra/pkg/util/zip"
)

// Onion is an interface for the layers of messages each encrypted inside a
// OnionSkin, which provides the cipher for the inner layers inside it.
type Onion interface {
	Magic() string
	Encode(s *zip.Splice) (e error)
	Decode(s *zip.Splice) (e error)
	Len() int
	Wrap(inner Onion)
	Handle(s *zip.Splice, p Onion, ng *Engine) (e error)
}

type Transport interface {
	Chain(t Transport) Transport
	Send(b slice.Bytes)
	Receive() <-chan slice.Bytes
}
