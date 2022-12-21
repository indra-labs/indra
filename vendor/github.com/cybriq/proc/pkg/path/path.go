package path

import (
	"strings"

	"github.com/cybriq/proc/pkg/util"
)

type Path []string

func (p Path) TrimPrefix() Path {
	if len(p) > 1 {
		return p[1:]
	}
	return p[:0]
}

func (p Path) String() string {
	return strings.Join(p, " ")
}

func From(s string) (p Path) {
	p = strings.Split(s, " ")
	return
}

func (p Path) Parent() (p1 Path) {
	if len(p) > 0 {
		p1 = p[:len(p)-1]
	}
	return
}

func (p Path) Child(child string) (p1 Path) {
	p1 = append(p, child)
	// log.I.Ln(p, p1)
	return
}

func (p Path) Common(p2 Path) (o Path) {
	for i := range p {
		if len(p2) < i {
			if p[i] == p2[i] {
				o = append(o, p[i])
			}
		}
	}
	return
}

func (p Path) Equal(p2 Path) bool {
	if len(p) == len(p2) {
		for i := range p {
			if util.Norm(p[i]) !=
				util.Norm(p2[i]) {
				return false
			}
		}
		return true
	}
	return false
}

func GetIndent(d int) string {
	return strings.Repeat("\t", d)
}
