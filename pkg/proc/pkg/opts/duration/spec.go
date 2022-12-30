package duration

import (
	"fmt"
	"strings"
	"time"

	"github.com/cybriq/proc/pkg/opts/config"
	"github.com/cybriq/proc/pkg/opts/meta"
	"github.com/cybriq/proc/pkg/path"
	"go.uber.org/atomic"
)

type Opt struct {
	p path.Path
	m meta.Metadata
	v atomic.Duration
	h []Hook
}

func (o *Opt) Path() (p path.Path) {
	return o.p
}

func (o *Opt) SetPath(p path.Path) {
	o.p = p
}

var _ config.Option = &Opt{}

type Hook func(*Opt)

func New(m meta.Data, h ...Hook) (o *Opt) {
	o = &Opt{m: meta.New(m, meta.Duration), h: h}
	_ = o.FromString(m.Default)
	return
}

func (o *Opt) Meta() meta.Metadata     { return o.m }
func (o *Opt) Type() meta.Type         { return o.m.Typ }
func (o *Opt) ToOption() config.Option { return o }

func (o *Opt) RunHooks() (e error) {
	for i := range o.h {
		o.h[i](o)
	}
	return
}

func (o *Opt) FromValue(v time.Duration) *Opt {
	o.v.Store(v)
	return o
}

func (o *Opt) FromString(s string) (e error) {
	s = strings.TrimSpace(s)
	var d time.Duration
	d, e = time.ParseDuration(s)
	if e != nil {
		return e
	}
	o.v.Store(d)
	e = o.RunHooks()
	return
}

func (o *Opt) String() (s string) {
	return fmt.Sprint(o.v.Load())
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
	c.Duration = func() time.Duration { return o.v.Load() }
	return
}

func Clamp(o *Opt, min, max time.Duration) func(*Opt) {
	return func(o *Opt) {
		v := o.v.Load()
		if v < min {
			o.v.Store(min)
		} else if v > max {
			o.v.Store(max)
		}
	}
}
