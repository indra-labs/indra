// Package path provides a simple string slice representation for paths, equally usable for filesystems or HD keychain schemes.
package path

import (
	"strings"

	log2 "git.indra-labs.org/dev/ind/pkg/proc/log"
	"git.indra-labs.org/dev/ind/pkg/util/norm"
)

type (
	Path []string
)

var (
	log   = log2.GetLogger()
	check = log.E.Chk
)

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

func From(s string) (p Path) { return strings.Split(s, " ") }

func GetIndent(d int) string { return strings.Repeat("\t", d) }

func (p Path) Parent() (p1 Path) {
	if len(p) > 0 {
		p1 = p[:len(p)-1]
	}
	return
}

func (p Path) String() string { return strings.Join(p, " ") }

func (p Path) TrimPrefix() Path {
	if len(p) > 1 {
		return p[1:]
	}
	return p[:0]
}
