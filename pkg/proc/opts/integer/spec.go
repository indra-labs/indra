package integer

import (
	"git.indra-labs.org/dev/ind/pkg/util/path"
	"strconv"
	"strings"

	"go.uber.org/atomic"

	log2 "git.indra-labs.org/dev/ind/pkg/proc/log"
	"git.indra-labs.org/dev/ind/pkg/proc/opts/config"
	"git.indra-labs.org/dev/ind/pkg/proc/opts/meta"
)

var (
	log   = log2.GetLogger()
	check = log.E.Chk
)

type Opt struct {
	p path.Path
	m meta.Metadata
	v atomic.Int64
	h []Hook
}

func (o *Opt) Path() (p path.Path) { return o.p }

func (o *Opt) SetPath(p path.Path) { o.p = p }

var _ config.Option = &Opt{}

type Hook func(*Opt) error

func New(m meta.Data, h ...Hook) (o *Opt) {
	o = &Opt{m: meta.New(m, meta.Integer), h: h}
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

func (o *Opt) FromValue(v int64) *Opt {
	o.v.Store(v)
	return o
}

func (o *Opt) FromString(s string) (e error) {
	s = strings.TrimSpace(s)
	var p int64
	if p, e = strconv.ParseInt(s, 10, 64); check(e) {
		return e
	}
	o.v.Store(p)
	e = o.RunHooks()
	return
}

func (o *Opt) String() (s string) {
	return strconv.FormatInt(o.v.Load(), 10)
}

func (o *Opt) Expanded() (s string) { return o.String() }

func (o *Opt) SetExpanded(s string) {
	err := o.FromString(s)
	check(err)
}

func (o *Opt) Value() (c config.Concrete) {
	c = config.NewConcrete()
	c.Integer = func() int64 { return o.v.Load() }
	return
}

func Clamp(o *Opt, min, max int64) func(*Opt) {
	return func(o *Opt) {
		v := o.v.Load()
		if v < min {
			o.v.Store(min)
		} else if v > max {
			o.v.Store(max)
		}
	}
}
