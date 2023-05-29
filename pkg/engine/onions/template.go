package onions

import (
	"git-indra.lan/indra-labs/indra/pkg/engine/coding"
	"git-indra.lan/indra-labs/indra/pkg/engine/sess"
	"git-indra.lan/indra-labs/indra/pkg/engine/sessions"
	"git-indra.lan/indra-labs/indra/pkg/util/splice"
)

const (
	EndMagic = "!!!!"
	EndLen   = 0
)

type End struct{}

func (x *End) Account(res *sess.Data, sm *sess.Manager,
	s *sessions.Data, last bool) (skip bool, sd *sessions.Data) {
	return
}

func (x *End) Decode(s *splice.Splice) (e error)                   { return }
func (x *End) Encode(s *splice.Splice) (e error)                   { return }
func EndGen() coding.Codec                                         { return &End{} }
func (x *End) GetOnion() interface{}                               { return x }
func (x *End) Handle(s *splice.Splice, p Onion, ni Ngin) (e error) { return }
func (x *End) Len() int                                            { return EndLen }
func (x *End) Magic() string                                       { return EndMagic }
func (x *End) Wrap(inner Onion)                                    {}
func NewEnd() *End                                                 { return &End{} }
func init()                                                        { Register(EndMagic, EndGen) }
