package engine

import (
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/engine/types"
	"git-indra.lan/indra-labs/indra/pkg/util/octet"
)

const (
	ConfirmationMagic = "cn"
	ConfirmationLen   = MagicLen + nonce.IDLen + 1
)

type Confirmation struct {
	nonce.ID
	Load byte
}

var confirmationPrototype types.Onion = &Confirmation{}

func init() { Register(ConfirmationMagic, confirmationPrototype) }

func (o Skins) Confirmation(id nonce.ID, load byte) Skins {
	return append(o, &Confirmation{ID: id, Load: load})
}

func (x *Confirmation) Magic() string { return ConfirmationMagic }

func (x *Confirmation) Encode(s *octet.Splice) (e error) {
	return s.Magic(ConfirmationMagic).ID(x.ID).Byte(x.Load)
}

func (x *Confirmation) Decode(s *octet.Splice) (e error) {
	if e = TooShort(s.Remaining(), ConfirmationLen-MagicLen,
		ConfirmationMagic); check(e) {
		return
	}
	return s.ReadID(&x.ID).ReadByte(&x.Load)
}

func (x *Confirmation) Len() int { return ConfirmationLen }

func (x *Confirmation) Wrap(inner types.Onion) {}

func (x *Confirmation) Handle(s *octet.Splice, p types.Onion,
	ng *Engine) (e error) {
	
	// When a confirmation arrives check if it is registered for and run the
	// hook that was registered with it.
	ng.PendingResponses.ProcessAndDelete(x.ID, nil)
	return
}
