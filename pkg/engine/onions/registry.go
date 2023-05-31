package onions

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/gookit/color"
	"github.com/indra-labs/indra/pkg/engine/coding"
	"github.com/indra-labs/indra/pkg/util/splice"
	"reflect"
	"sync"
)

var registry = NewRegistry()

type (
	CodecGenerators map[string]func() coding.Codec
	Registry        struct {
		sync.Mutex
		CodecGenerators
	}
)

func NewRegistry() *Registry {
	return &Registry{CodecGenerators: make(CodecGenerators)}
}

func MakeCodec(magic string) (cdc coding.Codec) {
	var in func() coding.Codec
	var ok bool
	if in, ok = registry.CodecGenerators[magic]; ok {
		cdc = in()
	}
	return
}

func Recognise(s *splice.Splice) (cdc coding.Codec) {
	registry.Lock()
	defer registry.Unlock()
	var magic string
	// log.D.S("splice", s.GetAll().ToBytes())
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

func Register(magicString string, on func() coding.Codec) {
	registry.Lock()
	defer registry.Unlock()
	registry.CodecGenerators[magicString] = on
}
