package wire

import (
	"fmt"
	"reflect"

	"github.com/Indra-Labs/indra"
	"github.com/Indra-Labs/indra/pkg/slice"
	. "github.com/cybriq/proc/pkg/log"
)

var (
	log   = GetLogger(indra.PathBase)
	check = log.E.Chk
)

// MagicLen is 3 to make it infeasible that the wrong cipher will yield a
// valid Magic string as listed below.
const MagicLen = 3

const ErrWrongMagic = "expected '%v', got '%v': type: %v"

func CheckMagic(b slice.Bytes, m slice.Bytes) (match bool) {
	header := b[:MagicLen]
	match = true
	for i := range header {
		if header[i] != m[i] {
			match = false
			break
		}
	}
	return
}

func ReturnError(errString string, t interface{}, found,
	magic slice.Bytes) (in interface{}, e error) {
	e = fmt.Errorf(errString,
		found[:MagicLen], magic, reflect.TypeOf(t))
	return

}
