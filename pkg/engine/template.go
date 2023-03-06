package engine

import (
	"git-indra.lan/indra-labs/indra/pkg/engine/types"
	"git-indra.lan/indra-labs/indra/pkg/relay/messages/magicbytes"
	"git-indra.lan/indra-labs/indra/pkg/util/octet"
)

const (
	TmplMagic = "!!"
	TmplLen   = magicbytes.Len
)

var TmplPrototype types.Onion = &Tmpl{}

type Tmpl struct{}

func NewTmpl() *Tmpl {
	return &Tmpl{}
}

func (x *Tmpl) Magic() string { return TmplMagic }

func (x *Tmpl) Encode(s *octet.Splice) (e error) {
	return s
}

func (x *Tmpl) Decode(s *octet.Splice) (e error) {
	if e = TooShort(s.Remaining(), TmplLen-MagicLen, TmplMagic); check(e) {
		return
	}
	return s
}

func (x *Tmpl) Len() int { return TmplLen }

func (x *Tmpl) Wrap(inner types.Onion) {}

func (x *Tmpl) Handle(s *octet.Splice, p types.Onion,
	ng *Engine) (e error) {
	
	return
}

func init() { Register(TmplMagic, TmplPrototype) }
