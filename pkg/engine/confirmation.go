package engine

import (
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/engine/magic"
	"git-indra.lan/indra-labs/indra/pkg/util/octet"
)

const (
	ConfirmationMagic = "cn"
	ConfirmationLen   = magic.Len + nonce.IDLen + 1
)

type Confirmation struct {
	nonce.ID
	Load byte
}

func confirmationPrototype() Onion { return &Confirmation{} }

func init() { Register(ConfirmationMagic, confirmationPrototype) }

func (o Skins) Confirmation(id nonce.ID, load byte) Skins {
	return append(o, &Confirmation{ID: id, Load: load})
}

func (x *Confirmation) Magic() string { return ConfirmationMagic }

func (x *Confirmation) Encode(s *octet.Splice) (e error) {
	s.Magic(ConfirmationMagic).ID(x.ID).Byte(x.Load)
	return
}

func (x *Confirmation) Decode(s *octet.Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), ConfirmationLen-magic.Len,
		ConfirmationMagic); check(e) {
		return
	}
	s.ReadID(&x.ID).ReadByte(&x.Load)
	return
}

func (x *Confirmation) Len() int { return ConfirmationLen }

func (x *Confirmation) Wrap(inner Onion) {}

func (x *Confirmation) Handle(s *octet.Splice, p Onion,
	ng *Engine) (e error) {
	
	// When a confirmation arrives check if it is registered for and run the
	// hook that was registered with it.
	ng.PendingResponses.ProcessAndDelete(x.ID, s.GetRange(-1, -1))
	return
}
