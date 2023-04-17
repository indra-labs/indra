package engine

import (
	"git-indra.lan/indra-labs/indra/pkg/engine/coding"
	"git-indra.lan/indra-labs/indra/pkg/engine/sessionmgr"
	"git-indra.lan/indra-labs/indra/pkg/engine/sessions"
	"git-indra.lan/indra-labs/indra/pkg/splice"
)

const (
	EndMagic = "!!"
	EndLen   = 0
)

func EndGen() coding.Codec           { return &End{} }
func init()                          { Register(EndMagic, EndGen) }
func (x *End) Magic() string         { return EndMagic }
func (x *End) Len() int              { return EndLen }
func (x *End) Wrap(inner Onion)      {}
func (x *End) GetOnion() interface{} { return x }

type End struct{}

func NewEnd() *End {
	return &End{}
}

func (x *End) Encode(s *splice.Splice) (e error) {
	return
}

func (x *End) Decode(s *splice.Splice) (e error) {
	return
}

func (x *End) Handle(s *splice.Splice, p Onion,
	ni interface{}) (e error) {
	
	return
}

func (x *End) Account(res *sessionmgr.Data, sm *sessionmgr.Manager,
	s *sessions.Data, last bool) (skip bool, sd *sessions.Data) {
	return
}
