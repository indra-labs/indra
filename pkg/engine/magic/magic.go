// Package magic is a simple specification and error helper for message identifying 4 byte strings that are used for the switching logic of a relay.
package magic

import "fmt"

const (
	// Len is the length in bytes of the magic bytes that prefixes all Indra
	// messages.
	Len = 4

	// ErrTooShort is an error for codec.Codec implementations to signal a message
	// buffer is shorter than the minimum defined for the message type.
	ErrTooShort = "'%s' message  minimum size: %d got: %d"
)

// TooShort is a helper function to return an error for a truncated packet.
func TooShort(got, found int, magic string) (e error) {
	if got >= found {
		return
	}
	e = fmt.Errorf(ErrTooShort, magic, found, got)
	return

}
