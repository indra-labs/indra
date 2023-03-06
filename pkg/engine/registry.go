package engine

import (
	"reflect"
	"sync"
	
	"git-indra.lan/indra-labs/indra/pkg/engine/types"
	"git-indra.lan/indra-labs/indra/pkg/util/octet"
)

var registry = NewRegistry()

type Onions map[string]types.Onion

type Registry struct {
	sync.Mutex
	Onions
}

func NewRegistry() *Registry {
	return &Registry{Onions: make(Onions)}
}

func Register(magicString string, on types.Onion) {
	registry.Lock()
	defer registry.Unlock()
	log.T.Ln("registering type", magicString, reflect.TypeOf(on))
	registry.Onions[magicString] = on
}

func Recognise(s *octet.Splice) (on types.Onion) {
	registry.Lock()
	defer registry.Unlock()
	var magic string
	if e := s.ReadMagic(&magic); check(e) {
		return
	}
	var ok bool
	var in types.Onion
	if in, ok = registry.Onions[magic]; ok {
		reflect.Copy(reflect.ValueOf(on), reflect.ValueOf(in))
	}
	return
}
