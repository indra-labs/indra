package toggle

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cybriq/proc/pkg/opts/config"
	"github.com/cybriq/proc/pkg/path"
	"go.uber.org/atomic"

	"github.com/cybriq/proc/pkg/opts/meta"
)

type Opt struct {
	p path.Path
	m meta.Metadata
	v atomic.Bool
	h []Hook
}

func (o *Opt) Path() (p path.Path) {
	return o.p
}

func (o *Opt) SetPath(p path.Path) {
	o.p = p
}

var _ config.Option = &Opt{}

type Hook func(*Opt) error

func New(m meta.Data, h ...Hook) (o *Opt) {
	m.Default = "false"
	o = &Opt{m: meta.New(m, meta.Bool), h: h}
	_ = o.FromString(m.Default)
	return
}

func (o *Opt) Meta() meta.Metadata     { return o.m }
func (o *Opt) Type() meta.Type         { return o.m.Typ }
func (o *Opt) ToOption() config.Option { return o }

func (o *Opt) RunHooks() (e error) {
	for i := range o.h {
		e = o.h[i](o)
		if e != nil {
			return
		}
	}
	return
}

func (o *Opt) FromValue(v bool) *Opt {
	o.v.Store(v)
	return o
}

func (o *Opt) FromString(s string) (e error) {
	s = strings.TrimSpace(s)
	switch s {
	case "f", "false", "off", "-":
		o.v.Store(false)
	case "t", "true", "on", "+":
		o.v.Store(true)
	default:
		return fmt.Errorf("string '%s' does not parse to boolean", s)
	}
	e = o.RunHooks()
	return
}

func (o *Opt) String() (s string) {
	return strconv.FormatBool(o.v.Load())
}

func (o *Opt) Expanded() (s string) {
	return o.String()
}

func (o *Opt) SetExpanded(s string) {
	err := o.FromString(s)
	log.E.Chk(err)
}

func (o *Opt) Value() (c config.Concrete) {
	c = config.NewConcrete()
	c.Bool = func() bool { return o.v.Load() }
	return
}
