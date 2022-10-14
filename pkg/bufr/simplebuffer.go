package bufr

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type Serializer interface {
	// Encode returns the wire/storage form of the data.
	Encode() []byte
	// Decode stores the decoded data from the head of the slice and returns
	// the remainder.
	Decode(b []byte) []byte
	// Len returns the length of the element in bytes.
	Len() int
}

type Serializers []Serializer

// Container is a storage format that automates conversion to bytes for
// arbitrary strings of data, and allows on demand partial decoding on the
// receiving side.
//
// Binary Format:
//
// Magic - 4 bytes - identifier for message
// Count - 2 bytes - number of elements in message
// Offsets - 4 bytes * Count - Starting index of each segment
// Rest of data is segmented as per Offsets
//
// With this, consuming code can automatically construct a packet by stringing
// together the correct sequence of fields, and on the receiving side,
// selectively decode segments as needed rather than preemptively decoding the
// entire packet.
type Container struct {
	Magic
	Data []byte
	Serializers
}

const (
	MagicLen    = 4
	CountLen    = 2
	OffsetLen   = 4
	OffsetStart = MagicLen + CountLen
)

type Magic [MagicLen]byte

// CreateContainer accepts a set of Serializers that assumes no data is to be
// encoded into the Data store, is intended to be called prior to LoadFromBytes.
//
// This function can become a factory by embedding into a closure returning
// function to create a specified message type.
func CreateContainer(spec Serializers, magic Magic) (c *Container) {
	c = &Container{Serializers: spec, Magic: magic}
	return
}

// LoadFromBytes accepts the raw bytes assumed to be encoded by the same
// Serializers and attempts to decode the values into a new container.
//
// This is the reverse process as CreateAndLoadContainer. Values can then be
// accessed via the Serializers slice. If the bytes are not a valid Container,
// the accessors will potentially panic.
func (c *Container) LoadFromBytes(b []byte) (e error) {
	if bytes.Compare(c.Magic[:], b[:MagicLen]) != 0 {
		e = fmt.Errorf("invalid magic, expected %v got %v",
			c.Magic, b[:MagicLen])
		return
	}
	c.Data = b
	return
}

// LoadFromSlice accepts a slice of interface and uses type switches to identify
// types for each segment and if they match or are convertible are loaded and
// the bytes are created to match it.
func (c *Container) LoadFromSlice(slice []interface{}) {

}

// CreateAndLoadContainer takes an array of serializer interface objects and
// renders the data into bytes
func (s Serializers) CreateAndLoadContainer(magic Magic) (out *Container) {
	out = &Container{Serializers: s}
	nodes := make([]uint32, len(s))
	// Total data length of container includes header, offset count and 4
	// bytes for each offset. This value represents the first offset.
	totalLen := len(s)*OffsetLen + CountLen + MagicLen
	for i := range s {
		nodes[i] = uint32(totalLen)
		totalLen += s[i].Len()
	}
	o := make([]byte, totalLen)
	// copy in magic bytes
	copy(o[:MagicLen], magic[:])
	// write in count value
	binary.LittleEndian.PutUint16(o[len(magic):len(magic)+2], uint16(len(s)))
	// write in offsets
	var end, start int
	for i := range nodes {
		start = OffsetStart + i*OffsetLen
		end = start + OffsetLen
		binary.LittleEndian.PutUint32(o[start:end], nodes[i])
	}
	// decode out the values
	for i := range s {
		start = end
		end += s[i].Len()
		copy(o[start:end], s[i].Encode())
	}
	out.Data = o
	return
}

func (c *Container) Count() uint16 {
	return binary.LittleEndian.Uint16(c.Data[MagicLen : MagicLen+CountLen])

}

func (c *Container) GetMagic() (out Magic) {
	copy(out[:], c.Data[:4])
	return
}

func (c *Container) GetOffset(idx uint16) (offset uint32, e error) {
	cnt := c.Count()
	if cnt < idx {
		e = fmt.Errorf("offset out of bounds - container "+
			"segment length %d", cnt)
		return
	}
	start := uint32(idx*OffsetLen + OffsetStart)
	end := offset + OffsetLen
	if len(c.Data) < int(end) {
		e = fmt.Errorf("container length shorter than offset %d, %d",
			end, len(c.Data))
		return
	}
	offset = binary.LittleEndian.Uint32(c.Data[start:end])
	return
}

// Get returns the bytes relevant to a specific segment of the message. By
// importing this segment into the data type corresponding to the segment the
// data can then be decoded and accessed without decoding the entire packet.
//
// The output of this function should be fed into a new instance of
func (c *Container) Get(idx uint16) (out []byte, e error) {
	length := c.Count()
	if length >= idx {
		var start, end uint32
		if start, e = c.GetOffset(idx); log.E.Chk(e) {
			return
		}
		if length > idx {
			if end, e = c.GetOffset(idx + 1); log.E.Chk(e) {
				return
			}
		} else {
			end = uint32(len(c.Data))
		}
		out = c.Data[start:end]
	} else {
		e = fmt.Errorf("index larger than container segment: %d > %d",
			idx, length)
	}
	return
}
