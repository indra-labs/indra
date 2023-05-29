package onions

import (
	"git-indra.lan/indra-labs/indra"
	"git-indra.lan/indra-labs/indra/pkg/engine/coding"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/util/splice"
	"github.com/gookit/color"
	"reflect"
	"sync"
)

var (
	log      = log2.GetLogger(indra.PathBase)
	fails    = log.E.Chk
	registry = NewRegistry()
)

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

func Recognise(s *splice.Splice) (cdc coding.Codec) {
	registry.Lock()
	defer registry.Unlock()
	var magic string
	// log.D.S("splice", s.GetAll().ToBytes())
	s.ReadMagic(&magic)
	var ok bool
	var in func() coding.Codec
	if in, ok = registry.CodecGenerators[magic]; ok {
		cdc = in()
	}
	if !ok || cdc == nil {
		// log.D.F("unrecognised magic %s ignoring message",
		// 	color.Red.Sprint(magic),
		// 	spew.Sdump(s.GetUntil(s.GetCursor()).ToBytes()),
		// 	spew.Sdump(s.GetFrom(s.GetCursor()).ToBytes()),
		// )
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
