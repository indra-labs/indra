package magicbytes

import (
	"fmt"
	"reflect"

	"github.com/Indra-Labs/indra"
	"github.com/Indra-Labs/indra/pkg/slice"
	log2 "github.com/cybriq/proc/pkg/log"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

const (
	// Len is 3 to make it infeasible that the wrong cipher will yield a
	// valid Magic string as listed below.
	Len           = 3
	ErrWrongMagic = "expected '%v', got '%v': type: %v"
	ErrTooShort   = "'%s' message  minimum size: %d got: %d"
)

func CheckMagic(b slice.Bytes, m slice.Bytes) (match bool) {
	header := b[:Len]
	match = true
	for i := range header {
		if header[i] != m[i] {
			match = false
			break
		}
	}
	return
}

func WrongMagic(t interface{}, found, magic slice.Bytes) (e error) {
	e = fmt.Errorf(ErrWrongMagic,
		found[:Len], magic, reflect.TypeOf(t))
	return

}
func TooShort(got, found int, magic string) (e error) {
	e = fmt.Errorf(ErrTooShort, magic, got, found)
	return

}
