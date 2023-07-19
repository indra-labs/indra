// Package delay provides an onion message type that allows a client to specify an arbitrary delay time before processing the rest of an onion message.
package delay

import (
	"github.com/indra-labs/indra/pkg/codec"
	"github.com/indra-labs/indra/pkg/codec/onion/cores/end"
	"github.com/indra-labs/indra/pkg/codec/ont"
	"github.com/indra-labs/indra/pkg/codec/reg"
	"github.com/indra-labs/indra/pkg/engine/magic"
	"github.com/indra-labs/indra/pkg/engine/sess"
	"github.com/indra-labs/indra/pkg/engine/sessions"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/util/slice"
	"github.com/indra-labs/indra/pkg/util/splice"
	"reflect"
	"time"
)

var (
	log   = log2.GetLogger()
	fails = log.E.Chk
)

const (
	Magic = "dely"
	Len   = magic.Len + slice.Uint64Len
)

// Delay is an instruction to hold a message for a specified time period before
// forwarding.
type Delay struct {
	time.Duration
	ont.Onion
}

// Account todo: record a decrement value based on a time coefficient of RelayRate.
func (x *Delay) Account(res *sess.Data, sm *sess.Manager,
	s *sessions.Data, last bool) (skip bool, sd *sessions.Data) {
	return
}

// Decode a splice.Splice's next bytes into a Delay.
func (x *Delay) Decode(s *splice.Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), Len-magic.Len, Magic); fails(e) {
		return
	}
	s.ReadDuration(&x.Duration)
	return
}

// Encode a Delay into a splice.Splice's next bytes.
func (x *Delay) Encode(s *splice.Splice) (e error) {
	log.T.S("encoding", reflect.TypeOf(x),
		x.Duration,
	)
	s.Magic(Magic).Uint64(uint64(x.Duration))
	if x.Onion != nil {
		e = x.Onion.Encode(s)
	}
	return
}

// Handle provides relay and accounting processing logic for receiving a Delay
// message.
func (x *Delay) Handle(s *splice.Splice, p ont.Onion, ng ont.Ngin) (e error) {
	// this is a message to hold the message in the buffer until a duration
	// elapses. The accounting for the remainder of the message adds a
	// factor to the effective byte consumption in accordance with the time
	// to be stored.
	// todo: accounting
	select {
	case <-time.After(x.Duration):
	}
	ng.HandleMessage(splice.BudgeUp(s), x)
	return
}

// Len returns the length of bytes required to encode this Delay.
func (x *Delay) Len() int { return Len + x.Onion.Len() }

// Magic bytes that identify this message.
func (x *Delay) Magic() string { return Magic }

// Wrap an Onion inside this Delay.
func (x *Delay) Wrap(inner ont.Onion) { x.Onion = inner }

// New creates a new Delay as an ont.Onion.
func New(d time.Duration) ont.Onion { return &Delay{Duration: d, Onion: &end.End{}} }

// Gen is a factory function for a Delay.
func Gen() codec.Codec { return &Delay{} }

func init() { reg.Register(Magic, Gen) }
