package list

import (
	"strings"

	"github.com/cybriq/proc/pkg/opts/config"
	"github.com/cybriq/proc/pkg/opts/meta"
	"github.com/cybriq/proc/pkg/opts/normalize"
	"github.com/cybriq/proc/pkg/path"
	"go.uber.org/atomic"
)

type Opt struct {
	p path.Path
	m meta.Metadata
	v atomic.Value
	x atomic.Value
	h []Hook
}

var _ config.Option = &Opt{}

func (o *Opt) Path() (p path.Path) {
	return o.p
}

func (o *Opt) SetPath(p path.Path) {
	o.p = p
}

type Hook func(*Opt) error

func New(m meta.Data, h ...Hook) (o *Opt) {
	o = &Opt{m: meta.New(m, meta.List), h: h}
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

func (o *Opt) FromValue(v []string) *Opt {
	o.v.Store(v)
	return o
}

func (o *Opt) FromString(s string) (e error) {
	s = strings.TrimSpace(s)
	split := strings.Split(s, ",")
	o.v.Store(split)
	e = o.RunHooks()
	return
}

func (o *Opt) String() (s string) {
	return strings.Join(o.v.Load().([]string), ",")
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
	c.List = func() []string { return o.v.Load().([]string) }
	return
}

// NormalizeNetworkAddress checks correctness of a network address
// specification, and adds a default path if needed, and enforces whether the
// port requires root permission and clamps it if not.
func NormalizeNetworkAddress(defaultPort string,
	userOnly bool) func(*Opt) error {

	return func(o *Opt) (e error) {
		var a []string
		a, e = normalize.Addresses(
			o.v.Load().([]string), defaultPort, userOnly)
		if !log.E.Chk(e) {
			o.x.Store(a)
		}
		return
	}
}

// NormalizeFilesystemPath cleans a directory specification, expands the ~ home
// folder shortcut, and if abs is set to true, returns the absolute path from
// filesystem root
func NormalizeFilesystemPath(abs bool, appName string) func(*Opt) error {
	return func(o *Opt) (e error) {
		strings := o.v.Load().([]string)
		for i := range strings {
			var cleaned string
			cleaned, e = normalize.ResolvePath(strings[i], appName, abs)
			if !log.E.Chk(e) {
				strings[i] = cleaned
			}
		}
		o.x.Store(strings)
		return
	}
}
