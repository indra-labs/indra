package onions

import (
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/engine/coding"
	"git-indra.lan/indra-labs/indra/pkg/engine/magic"
	"git-indra.lan/indra-labs/indra/pkg/engine/sess"
	"git-indra.lan/indra-labs/indra/pkg/engine/sessions"
	"git-indra.lan/indra-labs/indra/pkg/splice"
)

const (
	ConfirmationMagic = "conf"
	ConfirmationLen   = magic.Len + nonce.IDLen + 1
)

type Confirmation struct {
	ID   nonce.ID
	Load byte
}

func confirmationGen() coding.Codec           { return &Confirmation{} }
func init()                                   { Register(ConfirmationMagic, confirmationGen) }
func (x *Confirmation) Len() int              { return ConfirmationLen }
func (x *Confirmation) Wrap(inner Onion)      {}
func (x *Confirmation) GetOnion() interface{} { return x }

func (x *Confirmation) Magic() string { return ConfirmationMagic }

func (x *Confirmation) Encode(s *splice.Splice) (e error) {
	// log.T.S("encoding", reflect.TypeOf(x),
	// 	x.ID, x.Load,
	// )
	s.Magic(ConfirmationMagic).ID(x.ID).Byte(x.Load)
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

func (x *Confirmation) Handle(s *splice.Splice, p Onion, ng Ngin) (e error) {
	
	// When a confirmation arrives check if it is registered for and run the
	// hook that was registered with it.
	ng.Pending().ProcessAndDelete(x.ID, nil, s.GetAll())
	return
}

func (x *Confirmation) Account(res *sess.Data, sm *sess.Manager,
	s *sessions.Data, last bool) (skip bool, sd *sessions.Data) {
	
	res.ID = x.ID
	return
}
