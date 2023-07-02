// Package end is a null tombstone type onion message that indicates there is no more data in the onion (used with encoding only).
package end

import (
	"github.com/indra-labs/indra/pkg/engine/coding"
	"github.com/indra-labs/indra/pkg/engine/sess"
	"github.com/indra-labs/indra/pkg/engine/sessions"
	"github.com/indra-labs/indra/pkg/onions/ont"
	"github.com/indra-labs/indra/pkg/onions/reg"
	"github.com/indra-labs/indra/pkg/util/splice"
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

func (x *End) Decode(s *splice.Splice) (e error)                           { return }
func (x *End) Encode(s *splice.Splice) (e error)                           { return }
func EndGen() coding.Codec                                                 { return &End{} }
func (x *End) GetOnion() interface{}                                       { return x }
func (x *End) Handle(s *splice.Splice, p ont.Onion, ni ont.Ngin) (e error) { return }
func (x *End) Len() int                                                    { return EndLen }
func (x *End) Magic() string                                               { return EndMagic }
func (x *End) Wrap(inner ont.Onion)                                        {}
func NewEnd() *End                                                         { return &End{} }
func init()                                                                { reg.Register(EndMagic, EndGen) }
