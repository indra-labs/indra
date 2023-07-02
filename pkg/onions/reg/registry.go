// Package reg is a registry for message types that implement the coding.Codec interface.
//
// It is essentially a factory for messages, each registered message type has a function that returns a codec.Codec interface type containing the concrete type specified by the magic bytes.
//
// It can either be used to manually generate a type as in the factory model, or is used as a recogniser, accepting a splice.Splice and returning the concrete type indicated by the magic prefix of the splice.
package reg

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/gookit/color"
	"github.com/indra-labs/indra/pkg/engine/coding"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/util/splice"
	"reflect"
	"sync"
)

var (
	log   = log2.GetLogger()
	fails = log.E.Chk
)
var reg = newRegistry()

type (
	CodecGenerators map[string]func() coding.Codec
	registry        struct {
		sync.Mutex
		CodecGenerators
	}
)

func MakeCodec(magic string) (cdc coding.Codec) {
	var in func() coding.Codec
	var ok bool
	if in, ok = reg.CodecGenerators[magic]; ok {
		cdc = in()
	}
	return
}

func Recognise(s *splice.Splice) (cdc coding.Codec) {
	reg.Lock()
	defer reg.Unlock()
	var magic string
	s.ReadMagic(&magic)
	cdc = MakeCodec(magic)
	if cdc == nil {
		log.D.F("unrecognised magic %s ignoring message",
			color.Red.Sprint(magic),
			spew.Sdump(s.GetUntil(s.GetCursor()).ToBytes()),
			spew.Sdump(s.GetFrom(s.GetCursor()).ToBytes()),
		)
	} else {
		log.T.F("recognised magic %s for type %v",
			color.Red.Sprint(magic),
			color.Green.Sprint(reflect.TypeOf(cdc)))
	}
	return
}

func Register(magicString string, cdc func() coding.Codec) {
	reg.Lock()
	defer reg.Unlock()
	reg.CodecGenerators[magicString] = cdc
}

func newRegistry() *registry {
	return &registry{CodecGenerators: make(CodecGenerators)}
}
