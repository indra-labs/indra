package splice

import (
	"fmt"
	"reflect"
	"runtime"
)

// Bufferer is a simple interface to a raw bytes buffer with a cursor position.
type Bufferer interface {
	SetCursor(pos int)

	// GetCursor returns the current read/write position.
	GetCursor() (pos int)

	// Len returns the length required to encode the Splicer, *and* if already
	// allocated, the length that is allocated.
	Len() (l int)

	// GetBuffer returns the underlying raw bytes of a Splicer type.
	//
	// For the Splice type this means allocate the required length of buffer
	// for the message, if one does not exist.
	GetBuffer() (b *Buffer)
}

// Splicer is a generic, base interface type any type of data that can be strung
// together as a slice, each element being a field, that can be used to decode
// from binary form to runtime form and back.
type Splicer interface {
	Bufferer

	// Read decodes the next element, puts it inside the input interface, and
	// returns itself as the remainder. This enables chaining operations with
	// dot operators.
	Read(into Splicer) (self Splicer)

	// Write a Splicer into the buffer and return the splicer itself. This
	// enables chaining operations with dot operators.
	Write(from Splicer) (self Splicer)
}

// NoOp is a helper to trigger a panic for methods of a type that should not be
// called. This mainly is for the input type of Splices which does not come with
// a buffer, but can be itself encoded into binary form and back into
// runtime, if inserted into a Buffer.
func NoOp(typ, meth interface{}) {
	t := reflect.TypeOf(typ)
	m := runtime.FuncForPC(reflect.ValueOf(meth).Pointer()).Name()
	panic(fmt.Sprintf("method %s of type %s is a noop", m, t))

}

// Splices is a string of Splicer variables in their runtime form.
//
// This is the one type that only implements functions to compute length of
// the encoded bytes and to return a Buffer of such length.
type Splices []Splicer

// Len returns the total length required to encode the Splices.
func (s Splices) Len() (l int) {

	for i := range s {

		l += s[i].Len()
	}

	return
}

// GetBuffer allocates and returns a Buffer with the Splices inserted and a
// byte slice ready to encode with.
func (s Splices) GetBuffer() (b *Buffer) {

	b = &Buffer{
		Splices: s,
		Bytes:   make([]byte, s.Len()),
	}

	return
}

// The following implementations are noops and trigger a panic because there is
// no buffer to encode/decode in a Splices.
//
// The types are implemented primarily for computation of length of the Splicers
// inside as a concatenated byte slice.

func (s Splices) SetCursor(pos int) { NoOp(s, s.SetCursor) }
func (s Splices) GetCursor() (pos int) {
	NoOp(s, s.GetCursor)
	return
}
func (s Splices) Read(into Splicer) (self Splicer) {
	NoOp(s, s.Read)
	return
}
func (s Splices) Write(from Splicer) (self Splicer) {
	NoOp(s, s.Write)
	return
}

// Buffer is a raw byte slice that can have other Splicer data written into it,
// and reads back out a Splices.
//
// Buffer is a raw bytes field when embedded inside a Splice, thus usually only
// the last field (a separate, length prefixed Splicer is required for
// non-terminal elements), or the wrapper for a composition of Splicer fields..
//
// Buffer is the base Splicer type, which all member elements of Splices will
// embed.
type Buffer struct {
	Splices
	Bytes []byte
}

// NewBufferFrom creates a new Buffer based on a Splices, allocating the byte slice
// required to encode it.
//
// Variadic parameter enables construction from a function call, or an existing
// Splices can be unrolled using the ellipsis operator.
//
// This function will create a Buffer from any kind of Splicer or Splices.
func NewBufferFrom(s ...Splicer) (b *Buffer) {

	b = &Buffer{Splices: s, Bytes: make([]byte, Splices(s).Len())}

	return
}

// Read decodes and places the next field's runtime form into the `into`
// parameter, advances the cursor and returns itself, so it can be chained
// into a series of reads.
func (b *Buffer) Read(into Splicer) (self Splicer) {

	// todo: implement!
	return
}

// Write encodes the runtime value(s) in the `from` parameter into raw byte
// format, advances the cursor and returns itself.
func (b *Buffer) Write(from Splicer) (self Splicer) {

	// todo: implement!
	return
}
