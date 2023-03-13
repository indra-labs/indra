package engine

import (
	"git-indra.lan/indra-labs/indra/pkg/engine/magic"
	"git-indra.lan/indra-labs/indra/pkg/util/octet"
)

const (
	TmplMagic = "!!"
	TmplLen   = magic.Len
)

func TmplPrototype() Onion { return &Tmpl{} }

func init() { Register(TmplMagic, TmplPrototype) }

type Tmpl struct{}

func (o Skins) Tmpl() Skins {
	return append(o, &Tmpl{})
}

func NewTmpl() *Tmpl {
	return &Tmpl{}
}

func (x *Tmpl) Magic() string { return TmplMagic }

func (x *Tmpl) Encode(s *octet.Splice) (e error) {
	return
}

func (x *Tmpl) Decode(s *octet.Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), TmplLen-magic.Len, TmplMagic); check(e) {
		return
	}
	return
}

func (x *Tmpl) Len() int { return TmplLen }

func (x *Tmpl) Wrap(inner Onion) {}

func (x *Tmpl) Handle(s *octet.Splice, p Onion,
	ng *Engine) (e error) {
	
	return
}
