package splice

// Splice is a slice of Splicer - this is the data type representing an
// arbitrary sequence of different types that will be composed into a message.
//
// There is no methods defined directly on this type as it is simply an alias to
// embed this slice type in the following Splices type.
//
// Implementing the Splicer on this type would not make sense as that interface
// provides no logic for deciding which element to read or write. However,
// the Bufferer interface provides accessors, which are used by Splices.
type Splice []Splicer

func (s Splice) Len() (l int)            { return len(s) }
func (s Splice) GetBuffer() (b Bufferer) { return s }

// Splices is a Splicer combined with a Positioner.
//
// This provides a string of Splicers that can be iterated, by default when Read
// or Write are called, the cursor is incremented to the next position.
type Splices struct {
	pos int
	Splice
}

// NewSplices creates a new Splices.
//
// The zero values of this returned value indicate the writing position is at
// the beginning, and the length and underlying slice can be returned via the
// Bufferer implementations for the Splice.
//
// Because the Splice type provides a Bufferer interface, and is embedded in
// Splices, the Len and Getbuffer methods provide access to the slice-properties
// of the Splice.
func NewSplices() *Splices { return &Splices{} }

// Positioner implementations

func (s Splices) SetPos(pos int)    { s.pos = pos }
func (s Splices) GetPos() (pos int) { return s.pos }

// Splicer implementations

// Read copies the element at the cursor into a provided Splicer, using the
// Write method.
//
// The provided splicer, to which the element would be copied, is not modified
// if the position is at the end of the Splice.
func (s Splices) Read(into Splicer) (self Splicer) {
	
	// If the position is the same as the length it is after the last element
	// and there is no value to read.
	if s.pos < len(s.Splice) {
		
		// The Write method on a Splicer with no Positioner is equal to an
		// assignment operation, copying the value at the current position to
		// the single variable of the provided Splicer.
		//
		// If it is a SplicePositioner it will append the item from the current
		// item's cursor position to the end of the Splice.
		into.Write(s.Splice[s.pos])
		
		s.pos++
	}
	
	// There is no value if the cursor is at the end of the Splices.
	return s
}

// Write copies in a value from a Splicer's current position
func (s Splices) Write(from Splicer) (self Splicer) {
	
	// If the cursor is at the end, there is no element at the cursor.
	if s.pos == len(s.Splice) {
		return nil
	}
	
	// Place the provided Splicer into the current cursor position.
	s.Splice[s.pos] = from
	
	s.pos++
	
	return s
}
