package relay

import (
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/messages/intro"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

func (eng *Engine) SendIntro(id nonce.ID, target *Session, intr *intro.Layer,
	hook func(id nonce.ID, b slice.Bytes)) {
	
	hops := []byte{0, 1, 2, 3, 4, 5}
	s := make(Sessions, len(hops))
	s[2] = target
	se := eng.SelectHops(hops, s)
	var c Circuit
	copy(c[:], se)
	o := HiddenService(id, intr, se[len(se)-1], c, eng.KeySet)
	log.D.Ln("sending out intro onion")
	res := eng.PostAcctOnion(o)
	eng.SendWithOneHook(c[0].AddrPort, res, hook)
}
