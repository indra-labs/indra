package splice

// Bufferer is a simple interface to a raw bytes buffer with a cursor position.
// This interface provides cursor control, bytes buffer sizing and access to the
// buffer.
type Bufferer interface {
	
	// Len returns the length required to encode the Splicer, *and* if already
	// allocated, the length that is allocated.
	Len() (l int)
	
	// GetBuffer returns the underlying raw bytes of a Bufferer type.
	//
	// For the Splice type this means allocate the required length of buffer
	// for the message, if one does not exist.
	GetBuffer() (b Bufferer)
}

// Positioner is an interface for types that have a concept of a read and
// write position. This is combined with a Read/Write interface to enable a
// progressive encode/decode from raw bytes or could represent positions in
// any kind of slice or array.
type Positioner interface {
	
	// SetPos changes the current cursor position to the one provided, or
	// clamps it to the beginning or end if the position exceeds these
	// boundaries.
	SetPos(pos int)
	
	// GetPos returns the current read/write position.
	GetPos() (pos int)
}

// Splicer is a generic, base interface type any type of data that can be strung
// together as a slice, each element being a field, that can be used to decode
// from binary form to runtime form and back.
//
// Where the Splicer concrete type is a slice/array, these methods should
// enforce a rule that the Positioner type, if embedded with the Splicer, should
// increment the position after the read/write operation is done.
//
// The returning of the receiver value enables the chaining of these operations
// to implement structured types.
type Splicer interface {
	
	// Read decodes the next element, puts it inside the input interface, and
	// returns itself as the remainder. This enables chaining operations with
	// dot operators.
	Read(into Splicer) (self Splicer)
	
	// Write a Splicer into the buffer and return the splicer itself. This
	// enables chaining operations with dot operators.
	Write(from Splicer) (self Splicer)
}

// BufferPositioner is the hybrid of a Bufferer and a Positioner. This provides
// a uniform interface to accessing a slice type.
//
// This primarily will be used at a single simple interface for both slices of
// bytes (for encoding/decoding to/from wire) and slices of Splicers, thus
// forming the basis of the SpliceBuffer, which takes collections of Splicers
// and can encode and decode them from binary.
type BufferPositioner interface {
	Bufferer
	Positioner
}

// SplicePositioner is a hybrid Splicer and Positioner which can walk back and
// forth through a slice of Splicer variables, and the semantics of Read and
// Write methods is to overwrite if the position is already existing, append if
// at the end, and return the Splicer at the current cursor position.
//
// This type encompasses the input variable lists that are used to create a
// composite, serial form of the list of Splicer variables in their runtime
// form.
type SplicePositioner interface {
	Positioner
	Splicer
}

// SpliceBuffer is an interface for variables that contain a buffer,
// a cursor to indicate where to read and write out of the buffer,
// and a slice of in-process values that are either encoded from or decoded
// into from the Bufferer.
//
// It extends the SplicePositioner by providing a buffer that can be
// progressively decoded, or encoded, and at the end, the buffer sent over a
// wire to disk or across a network.
type SpliceBuffer interface {
	SplicePositioner
	Bufferer
}
