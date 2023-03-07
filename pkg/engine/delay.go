package engine

import (
	"time"
	
	"git-indra.lan/indra-labs/indra/pkg/util/octet"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

const (
	DelayMagic = "dl"
	DelayLen   = MagicLen + slice.Uint64Len
)

type Delay struct {
	time.Duration
	Onion
}

func delayPrototype() Onion { return &Delay{} }

func init() { Register(DelayMagic, delayPrototype) }

func (o Skins) Delay(d time.Duration) Skins {
	return append(o, &Delay{Duration: d, Onion: nop})
}

func (x *Delay) Magic() string { return DelayMagic }

func (x *Delay) Encode(s *octet.Splice) (e error) {
	return x.Onion.Encode(s.Magic(DelayMagic).Uint64(uint64(x.Duration)))
}

func (x *Delay) Decode(s *octet.Splice) (e error) {
	if e = TooShort(s.Remaining(), DelayLen-MagicLen, DelayMagic); check(e) {
		return
	}
	s.ReadDuration(&x.Duration)
	return
}

func (x *Delay) Len() int { return DelayLen + x.Onion.Len() }

func (x *Delay) Wrap(inner Onion) { x.Onion = inner }

func (x *Delay) Handle(s *octet.Splice, p Onion, ng *Engine) (e error) {
	
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
