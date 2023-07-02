// Package magic is a simple specification and error helper for message identifying 4 byte strings that are used for the switching logic of a relay.
package magic

import "fmt"

const (
	Len         = 4
	ErrTooShort = "'%s' message  minimum size: %d got: %d"
)

// TooShort is a helper function to return an error for a truncated packet.
func TooShort(got, found int, magic string) (e error) {
	if got >= found {
		return
	}
	e = fmt.Errorf(ErrTooShort, magic, got, found)
	return

}
