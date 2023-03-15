package engine

import (
	"git-indra.lan/indra-labs/indra/pkg/engine/magic"
	"git-indra.lan/indra-labs/indra/pkg/util/octet"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

const (
	TripleMagic = "rv"
	TripleLen   = ReverseHeaderLen
)

// func TriplePrototype() Onion { return &Triple{} }

// func init() { Register(TripleMagic, TriplePrototype) }

type Triple struct {
	slice.Bytes
	Onion
}

func (o Skins) Triple(header slice.Bytes) Skins {
	return append(o, &Triple{Bytes: header})
}

func (x *Triple) Magic() string { return TripleMagic }

// Encode the header, this is the only method that will be called, when the node
// receives it, it will not register this type as that has been commented out
// above.
func (x *Triple) Encode(s *octet.Splice) (e error) {
	s.GetRange(s.GetCursor(), s.Advance(TripleLen))
	return
}

func (x *Triple) Decode(s *octet.Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), TripleLen-magic.Len,
		TripleMagic); check(e) {
		return
	}
	return
}

func (x *Triple) Len() int { return TripleLen }

func (x *Triple) Wrap(inner Onion) { x.Onion = inner }

func (x *Triple) Handle(s *octet.Splice, p Onion,
	ng *Engine) (e error) {
	
	return
}
