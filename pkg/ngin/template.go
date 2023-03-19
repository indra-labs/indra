package ngin

import (
	"git-indra.lan/indra-labs/indra/pkg/ngin/magic"
	"git-indra.lan/indra-labs/indra/pkg/util/zip"
)

const (
	TmplMagic = "!!"
	TmplLen   = 0
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

func (x *Tmpl) Encode(s *zip.Splice) (e error) {
	return
}

func (x *Tmpl) Decode(s *zip.Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), TmplLen-magic.Len,
		TmplMagic); check(e) {
		return
	}
	return
}

func (x *Tmpl) Len() int { return TmplLen }

func (x *Tmpl) Wrap(inner Onion) {}

func (x *Tmpl) Handle(s *zip.Splice, p Onion,
	ng *Engine) (e error) {
	
	return
}
