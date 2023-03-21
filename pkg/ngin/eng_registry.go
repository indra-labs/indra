package ngin

import (
	"net/netip"
	"reflect"
	"sync"
	
	"github.com/gookit/color"
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

func Recognise(s *Splice, addr *netip.AddrPort) (on Onion) {
	registry.Lock()
	defer registry.Unlock()
	var magic string
	s.ReadMagic(&magic)
	var ok bool
	var in func() Onion
	if in, ok = registry.Onions[magic]; ok {
		on = in()
	}
	log.D.F("%s recognised magic %s for type %v",
		color.Yellow.Sprint(addr.String()), color.Red.Sprint(magic),
		color.Green.Sprint(reflect.TypeOf(on)))
	return
}
