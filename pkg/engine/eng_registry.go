package engine

import (
	"reflect"
	"sync"
	
	"github.com/gookit/color"
	
	"git-indra.lan/indra-labs/indra/pkg/engine/coding"
	"git-indra.lan/indra-labs/indra/pkg/splice"
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
	s.ReadMagic(&magic)
	var ok bool
	var in func() coding.Codec
	if in, ok = registry.CodecGenerators[magic]; ok {
		cdc = in()
	}
	if !ok {
		log.D.S("decryption failure", s.GetRest())
	}
	log.D.F("recognised magic %s for type %v",
		color.Red.Sprint(magic),
		color.Green.Sprint(reflect.TypeOf(cdc)))
	return
}
