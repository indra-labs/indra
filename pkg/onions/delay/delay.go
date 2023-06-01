package delay

import (
	"github.com/indra-labs/indra"
	"github.com/indra-labs/indra/pkg/engine/coding"
	"github.com/indra-labs/indra/pkg/engine/magic"
	"github.com/indra-labs/indra/pkg/engine/sess"
	"github.com/indra-labs/indra/pkg/engine/sessions"
	"github.com/indra-labs/indra/pkg/onions/end"
	"github.com/indra-labs/indra/pkg/onions/ont"
	"github.com/indra-labs/indra/pkg/onions/reg"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/util/slice"
	"github.com/indra-labs/indra/pkg/util/splice"
	"reflect"
	"time"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	fails = log.E.Chk
)

const (
	DelayMagic = "dely"
	DelayLen   = magic.Len + slice.Uint64Len
)

type Delay struct {
	time.Duration
	ont.Onion
}

func (x *Delay) Account(res *sess.Data, sm *sess.Manager,
	s *sessions.Data, last bool) (skip bool, sd *sessions.Data) {
	return
}

func (x *Delay) Decode(s *splice.Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), DelayLen-magic.Len, DelayMagic); fails(e) {
		return
	}
	s.ReadDuration(&x.Duration)
	return
}

func (x *Delay) Encode(s *splice.Splice) (e error) {
	log.T.S("encoding", reflect.TypeOf(x),
		x.Duration,
	)
	s.Magic(DelayMagic).Uint64(uint64(x.Duration))
	if x.Onion != nil {
		e = x.Onion.Encode(s)
	}
	return
}

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

func (x *Delay) Len() int                { return DelayLen + x.Onion.Len() }
func (x *Delay) Magic() string           { return DelayMagic }
func (x *Delay) Wrap(inner ont.Onion)    { x.Onion = inner }
func NewDelay(d time.Duration) ont.Onion { return &Delay{Duration: d, Onion: &end.End{}} }
func delayGen() coding.Codec             { return &Delay{} }
func init()                              { reg.Register(DelayMagic, delayGen) }
