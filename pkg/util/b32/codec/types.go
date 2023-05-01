package codec

import (
	"git-indra.lan/indra-labs/indra/pkg/util/b32/codecer"
)

// Codec is the collection of elements that creates a Human Readable Binary
// Transcription Codec
//
// This is an example of the use of a structure definition to encapsulate and
// logically connect together all of the elements of an implementation, while
// also permitting this to be used by external code without further
// dependencies, either through this type, or via the interface defined further
// down.
//
// It is not "official" idiom, but it's the opinion of the author of this
// tutorial that return values given in type specifications like this helps the
// users of the library understand what the return values actually are.
// Otherwise, the programmer is forced to read the whole function just to spot
// the names and, even worse, comments explaining what the values are, which are
// often neglected during debugging, and turn into lies!
type Codec struct {
	
	// Name is the human readable name given to this encoder
	Name string
	
	// HRP is the Human Readable Prefix to be appended in front of the encoding
	// to disambiguate it from another encoding or as a network or protocol
	// identifier. This can be empty, but more usually this will be used to
	// disambiguate versus other similarly encoded values, such as used on a
	// different cryptocurrency network, or between main and test networks.
	HRP string
	
	// Charset is the set of characters that the encoder uses. This should match
	// the output encoder, 32 for using base32, 64 for base64, etc.
	//
	// For arbitrary bases, see the following function in the standard library:
	// https://cs.opensource.google/go/go/+/refs/tags/go1.17.7:src/strconv/itoa.go;l=25
	// This function can render up to base36, but by default uses 0-9a-z in its
	// representation, which would either need to be string substituted for
	// non-performance-critical uses or the function above forked to provide a
	// direct encoding to the intended characters used for the encoding, using
	// this charset string as the key. The sequence matters, each character
	// represents the cipher for a given value to be found at a given place in
	// the encoded number.
	Charset string
	
	// Encode takes an arbitrary length byte input and returns the output as
	// defined for the codec
	Encoder func(input []byte) (output string, err error)
	
	// Decode takes an encoded string and returns if the encoding is valid and
	// the value passes any check function defined for the type.
	Decoder func(input string) (output []byte, err error)
	
	// AddCheck is used by Encode to add extra bytes for the checksum to ensure
	// correct input so user does not send to a wrong address by mistake, for
	// example.
	MakeCheck func(input []byte, checkLen int) (output []byte)
	
	// Check returns whether the check is valid
	Check func(input []byte) (err error)
}

// The following implementations are here to ensure this type implements the
// interface. In this tutorial/example we are creating a kind of generic
// implementation through the use of closures loaded into a struct.
//
// Normally a developer would use either one, or the other, a struct with
// closures, OR an interface with arbitrary variable with implementations for
// the created type.
//
// In order to illustrate both interfaces and the use of closures with a struct
// in this way we combine the two things by invoking the closures in a
// predefined pair of methods that satisfy the interface.
//
// In fact, there is no real reason why this design could not be standard idiom,
// since satisfies most of the requirements of idiom for both interfaces
// (minimal) and hot-reloadable interfaces (allowing creation of registerable
// compile time plugins such as used in database drivers with structs, and the
// end user can then either use interfaces or the provided struct, and both
// options are open.

// This ensures the interface is satisfied for codecer.Codecer and is removed in
// the generated binary because the underscore indicates the value is discarded.
var _ codecer.Codecer = &Codec{}

// Encode implements the codecer.Codecer.Encode by calling the provided
// function, and allows the concrete Codec type to always satisfy the interface,
// while allowing it to be implemented entirely differently.
//
// Note: short functions like this can be one-liners according to gofmt.
func (c *Codec) Encode(input []byte) (string, error) { return c.Encoder(input) }

// Decode implements the codecer.Codecer.Decode by calling the provided
// function, and allows the concrete Codec type to always satisfy the interface,
// while allowing it to be implemented entirely differently.
//
// Note: this also can be a one liner. Since we name the return values in the
// type definition and interface, omitting them here makes the line short enough
// to be a one liner.
func (c *Codec) Decode(input string) ([]byte, error) { return c.Decoder(input) }
