package onions

import (
	"reflect"
	"sync"
	
	"github.com/gookit/color"
	
	"git-indra.lan/indra-labs/indra"
	"git-indra.lan/indra-labs/indra/pkg/engine/coding"
	magic2 "git-indra.lan/indra-labs/indra/pkg/engine/magic"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/splice"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	fails = log.E.Chk
)

var registry = NewRegistry()

type CodecGenerators map[string]func() coding.Codec

type Registry struct {
	sync.Mutex
	CodecGenerators
}

func NewRegistry() *Registry {
	return &Registry{CodecGenerators: make(CodecGenerators)}
}

func Register(magicString string, on func() coding.Codec) {
	registry.Lock()
	defer registry.Unlock()
	registry.CodecGenerators[magicString] = on
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
	if !ok {
		log.D.S("message unrecognised", s.GetRange(s.GetCursor()-magic2.Len,
			-1).ToBytes())
	}
	log.D.F("recognised magic %s for type %v",
		color.Red.Sprint(magic),
		color.Green.Sprint(reflect.TypeOf(cdc)))
	return
}
