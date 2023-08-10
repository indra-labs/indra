// Package end is a null tombstone type onion message that indicates there is no more data in the onion (used with encoding only).
package end

import (
	"git.indra-labs.org/dev/ind/pkg/codec"
	"git.indra-labs.org/dev/ind/pkg/codec/ont"
	"git.indra-labs.org/dev/ind/pkg/codec/reg"
	"git.indra-labs.org/dev/ind/pkg/engine/sess"
	"git.indra-labs.org/dev/ind/pkg/engine/sessions"
	"git.indra-labs.org/dev/ind/pkg/util/splice"
)

const (
	Magic = "!!!!"
	Len   = 0
)

type End struct{}

func (x *End) Account(res *sess.Data, sm *sess.Manager,
	s *sessions.Data, last bool) (skip bool, sd *sessions.Data) {
	return
}

func (x *End) Decode(s *splice.Splice) (e error)                           { return }
func (x *End) Encode(s *splice.Splice) (e error)                           { return }
func EndGen() codec.Codec                                                  { return &End{} }
func (x *End) Unwrap() interface{}                                         { return x }
func (x *End) Handle(s *splice.Splice, p ont.Onion, ni ont.Ngin) (e error) { return }
func (x *End) Len() int {

	codec.MustNotBeNil(x)

	return Len
}
func (x *End) Magic() string        { return Magic }
func (x *End) Wrap(inner ont.Onion) {}
func NewEnd() *End                  { return &End{} }
func init()                         { reg.Register(Magic, EndGen) }
