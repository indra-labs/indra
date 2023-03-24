package engine

import (
	"reflect"
	"time"
	
	"git-indra.lan/indra-labs/indra/pkg/engine/magic"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

const (
	DelayMagic = "dl"
	DelayLen   = magic.Len + slice.Uint64Len
)

type Delay struct {
	time.Duration
	Onion
}

func delayPrototype() Onion       { return &Delay{} }
func init()                       { Register(DelayMagic, delayPrototype) }
func (x *Delay) Magic() string    { return DelayMagic }
func (x *Delay) Len() int         { return DelayLen + x.Onion.Len() }
func (x *Delay) Wrap(inner Onion) { x.Onion = inner }

func (o Skins) Delay(d time.Duration) Skins {
	return append(o, &Delay{Duration: d, Onion: nop})
}

func (x *Delay) Encode(s *Splice) (e error) {
	log.T.S("encoding", reflect.TypeOf(x),
		x.Duration,
	)
	s.
		Magic(DelayMagic).
		Uint64(uint64(x.Duration))
	if x.Onion != nil {
		e = x.Onion.Encode(s)
	}
	return
}

func (x *Delay) Decode(s *Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), DelayLen-magic.Len, DelayMagic); check(e) {
		return
	}
	s.ReadDuration(&x.Duration)
	return
}

func (x *Delay) Handle(s *Splice, p Onion, ng *Engine) (e error) {
	
	// this is a message to hold the message in the buffer until a duration
	// elapses. The accounting for the remainder of the message adds a
	// factor to the effective byte consumption in accordance with the time
	// to be stored.
	// todo: accounting
	select {
	case <-time.After(x.Duration):
	}
	ng.HandleMessage(BudgeUp(s), x)
	return
}
