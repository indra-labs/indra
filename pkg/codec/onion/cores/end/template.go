// Package end is a null tombstone type onion message that indicates there is no more data in the onion (used with encoding only).
package end

import (
	"github.com/indra-labs/indra/pkg/codec"
	"github.com/indra-labs/indra/pkg/codec/ont"
	"github.com/indra-labs/indra/pkg/codec/reg"
	"github.com/indra-labs/indra/pkg/engine/sess"
	"github.com/indra-labs/indra/pkg/engine/sessions"
	"github.com/indra-labs/indra/pkg/util/splice"
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
