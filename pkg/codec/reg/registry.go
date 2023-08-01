// Package reg is a registry for message types that implement the coding.Codec interface.
//
// It is essentially a factory for messages, each registered message type has a function that returns a codec.Codec interface type containing the concrete type specified by the magic bytes.
//
// It can either be used to manually generate a type as in the factory model, or is used as a recogniser, accepting a splice.Splice and returning the concrete type indicated by the magic prefix of the splice.
package reg

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/gookit/color"
	"github.com/indra-labs/indra/pkg/codec"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/util/splice"
	"sync"
)

var (
	log   = log2.GetLogger()
	fails = log.E.Chk
)
var reg = newRegistry()

type (
	// CodecGenerators is the map of generators that have been registered.
	CodecGenerators map[string]func() codec.Codec

	// registry is a factory constructor registry used with Magic bytes to enable
	// decoding of encoded message onions.
	registry struct {
		sync.Mutex
		CodecGenerators
	}
)

// MakeCodec is a simple function to find a Codec factory function and return its
// output.
func MakeCodec(magic string) (cdc codec.Codec) {
	var in func() codec.Codec
	var ok bool
	if in, ok = reg.CodecGenerators[magic]; ok {
		cdc = in()
	}
	return
}

// Recognise selects a factory function and steps the cursor forward after the
// Magic to be ready to call the Decode method on the Codec inside.
func Recognise(s *splice.Splice) (cdc codec.Codec) {
	reg.Lock()
	defer reg.Unlock()
	var magic string
	s.ReadMagic(&magic)
	cdc = MakeCodec(magic)
	if cdc == nil {
		log.D.F("unrecognised magic %s ignoring message",
			color.Red.Sprint(magic),
			spew.Sdump(s.GetUntil(s.GetCursor()).ToBytes()),
			spew.Sdump(s.GetFrom(s.GetCursor()).ToBytes()),
		)
		// } else {
		// 	log.T.F("recognised magic %s for type %v",
		// 		color.Red.Sprint(magic),
		// 		color.Green.Sprint(reflect.TypeOf(cdc)))
	}
	return
}

// Register adds a new factory constructor function associated with a provided Magic bytes.
func Register(magicString string, cdc func() codec.Codec) {
	reg.Lock()
	defer reg.Unlock()
	reg.CodecGenerators[magicString] = cdc
}

func newRegistry() *registry {
	return &registry{CodecGenerators: make(CodecGenerators)}
}
