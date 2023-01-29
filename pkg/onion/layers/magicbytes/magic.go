package magicbytes

import (
	"fmt"

	"git-indra.lan/indra-labs/indra"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

const (
	Len         = 2
	ErrTooShort = "'%s' message  minimum size: %d got: %d"
)

func TooShort(got, found int, magic string) (e error) {
	e = fmt.Errorf(ErrTooShort, magic, got, found)
	check(e)
	return

}
