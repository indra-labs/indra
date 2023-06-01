package confirmation

import (
	"github.com/indra-labs/indra"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/engine/coding"
	"github.com/indra-labs/indra/pkg/engine/magic"
	"github.com/indra-labs/indra/pkg/engine/sess"
	"github.com/indra-labs/indra/pkg/engine/sessions"
	"github.com/indra-labs/indra/pkg/onions/ont"
	"github.com/indra-labs/indra/pkg/onions/reg"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/util/splice"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	fails = log.E.Chk
)

const (
	ConfirmationMagic = "conf"
	ConfirmationLen   = magic.Len + nonce.IDLen + 1
)

type Confirmation struct {
	ID   nonce.ID
	Load byte
}

func (x *Confirmation) Account(res *sess.Data, sm *sess.Manager,
	s *sessions.Data, last bool) (skip bool, sd *sessions.Data) {
	res.ID = x.ID
	return
}

func (x *Confirmation) Decode(s *splice.Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), ConfirmationLen-magic.Len,
		ConfirmationMagic); fails(e) {
		return
	}
	s.ReadID(&x.ID).ReadByte(&x.Load)
	return
}

func (x *Confirmation) Encode(s *splice.Splice) (e error) {
	s.Magic(ConfirmationMagic).ID(x.ID).Byte(x.Load)
	return
}

func (x *Confirmation) GetOnion() interface{} { return x }

func (x *Confirmation) Handle(s *splice.Splice, p ont.Onion, ng ont.Ngin) (e error) {
	// When a confirmation arrives check if it is registered for and run the
	// hook that was registered with it.
	ng.Pending().ProcessAndDelete(x.ID, nil, s.GetAll())
	return
}

func (x *Confirmation) Len() int                       { return ConfirmationLen }
func (x *Confirmation) Magic() string                  { return ConfirmationMagic }
func (x *Confirmation) Wrap(inner ont.Onion)           {}
func NewConfirmation(id nonce.ID, load byte) ont.Onion { return &Confirmation{ID: id, Load: load} }
func confirmationGen() coding.Codec                    { return &Confirmation{} }
func init()                                            { reg.Register(ConfirmationMagic, confirmationGen) }
