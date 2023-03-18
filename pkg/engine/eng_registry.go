package engine

import (
	"reflect"
	"sync"
	
	"github.com/gookit/color"
	
	"git-indra.lan/indra-labs/indra/pkg/util/zip"
)

var registry = NewRegistry()

type Onions map[string]func() Onion

type Registry struct {
	sync.Mutex
	Onions
}

func NewRegistry() *Registry {
	return &Registry{Onions: make(Onions)}
}

func Register(magicString string, on func() Onion) {
	registry.Lock()
	defer registry.Unlock()
	registry.Onions[magicString] = on
}

func Recognise(s *zip.Splice) (on Onion) {
	registry.Lock()
	defer registry.Unlock()
	var magic string
	s.ReadMagic(&magic)
	var ok bool
	var in func() Onion
	if in, ok = registry.Onions[magic]; ok {
		on = in()
	}
	log.D.F("recognised magic %s for type %v", color.Red.Sprint(magic),
		color.Green.Sprint(reflect.TypeOf(on)))
	// log.T.S("recognition", s.GetRange(s.GetCursor()-magic2.Len, -1).ToBytes())
	return
}
