package engine

import (
	"reflect"
	"sync"
	
	"git-indra.lan/indra-labs/indra/pkg/util/octet"
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
	log.I.Ln("registering message type", magicString, reflect.TypeOf(on()))
	registry.Onions[magicString] = on
}

func Recognise(s *octet.Splice) (on Onion) {
	registry.Lock()
	defer registry.Unlock()
	var magic string
	s.ReadMagic(&magic)
	var ok bool
	var in func() Onion
	if in, ok = registry.Onions[magic]; ok {
		on = in()
	}
	return
}
