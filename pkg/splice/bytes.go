package splice

// Bytes is a simple alias for a byte slice that provides a Bufferer.
type Bytes []byte

func (b Bytes) Len() (l int)              { return len(b) }
func (b Bytes) GetBuffer() (buf Bufferer) { return b }

// ByteSlice is a simple byte slice with a cursor.
type ByteSlice struct {
	b   Bytes
	pos int
}

func NewByteSlice() *ByteSlice {
	return &ByteSlice{}
}

func (b ByteSlice) Len() (l int)              { return b.b.Len() }
func (b ByteSlice) GetBuffer() (buf Bufferer) { return b.b.GetBuffer() }
func (b ByteSlice) GetPos() (pos int)         { return b.pos }

func (b ByteSlice) SetPos(pos int) {
	
	// clamp the input position to the valid range for the buffer.
	if pos < 0 {
		pos = 0
		// If this condition is true the next one can't be.
	} else if pos > b.b.Len() {
		pos = b.b.Len()
	}
	b.pos = pos
}
