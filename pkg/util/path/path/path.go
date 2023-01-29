package path

import (
	"strings"

	"git-indra.lan/indra-labs/indra"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/util/norm"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

type Path []string

func (p Path) TrimPrefix() Path {
	if len(p) > 1 {
		return p[1:]
	}
	return p[:0]
}

func (p Path) String() string { return strings.Join(p, " ") }

func From(s string) (p Path) { return strings.Split(s, " ") }

func (p Path) Parent() (p1 Path) {
	if len(p) > 0 {
		p1 = p[:len(p)-1]
	}
	return
}

func (p Path) Child(child string) (p1 Path) { return append(p, child) }

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
			if norm.Norm(p[i]) !=
				norm.Norm(p2[i]) {
				return false
			}
		}
		return true
	}
	return false
}

func GetIndent(d int) string { return strings.Repeat("\t", d) }
