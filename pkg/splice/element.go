package splice

// Element is a single Splicer, which wraps a Splicer interface into a pointer
// to a struct. This type implements Read and Write, which function as assign
// and write their containing value into teh provided Splicer.
type Element struct {
	v Splicer
}

func NewElement(in Splicer) *Element { return &Element{v: in} }

// Read an Element into the provided Splicer (Write to provided Splicer,
// read from this Element.
func (e Element) Read(into Splicer) (self Splicer) {
	into.Write(e.v)
	return e.v
}

// Write a provided Splicer into the slot in this Element.
func (e Element) Write(from Splicer) (self Splicer) {
	e.v = from
	return e
}
