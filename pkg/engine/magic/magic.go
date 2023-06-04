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
