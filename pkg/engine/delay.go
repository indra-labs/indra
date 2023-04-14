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
	Mung
}

func delayGen() Codec            { return &Delay{} }
func init()                      { Register(DelayMagic, delayGen) }
func (x *Delay) Magic() string   { return DelayMagic }
func (x *Delay) Len() int        { return DelayLen + x.Mung.Len() }
func (x *Delay) Wrap(inner Mung) { x.Mung = inner }

func (o Skins) Delay(d time.Duration) Skins {
	return append(o, &Delay{Duration: d, Mung: nop})
}

func (x *Delay) Encode(s *Splice) (e error) {
	log.T.S("encoding", reflect.TypeOf(x),
		x.Duration,
	)
	s.Magic(DelayMagic).Uint64(uint64(x.Duration))
	if x.Mung != nil {
		e = x.Mung.Encode(s)
	}
	return
}

func (x *Delay) Decode(s *Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), DelayLen-magic.Len, DelayMagic); fails(e) {
		return
	}
	s.ReadDuration(&x.Duration)
	return
}

func (x *Delay) Handle(s *Splice, p Mung, ng *Engine) (e error) {
	
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

func (x *Delay) Account(res *SendData, sm *SessionManager, s *SessionData, last bool) (skip bool, sd *SessionData) {
	return
}
