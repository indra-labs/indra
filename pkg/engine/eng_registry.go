package engine

import (
	"net/netip"
	"reflect"
	"sync"
	
	"github.com/gookit/color"
)

var registry = NewRegistry()

type OnionGenerators map[string]func() Onion

type Registry struct {
	sync.Mutex
	OnionGenerators
}

func NewRegistry() *Registry {
	return &Registry{OnionGenerators: make(OnionGenerators)}
}

func Register(magicString string, on func() Onion) {
	registry.Lock()
	defer registry.Unlock()
	registry.OnionGenerators[magicString] = on
}

func Recognise(s *Splice, addr *netip.AddrPort) (on Onion) {
	registry.Lock()
	defer registry.Unlock()
	var magic string
	s.ReadMagic(&magic)
	var ok bool
	var in func() Onion
	if in, ok = registry.OnionGenerators[magic]; ok {
		on = in()
	}
	if !ok {
		log.D.S("decryption failure", s.GetCursorToEnd())
	}
	log.D.F("%s recognised magic %s for type %v",
		color.Yellow.Sprint(addr.String()), color.Red.Sprint(magic),
		color.Green.Sprint(reflect.TypeOf(on)))
	return
}
