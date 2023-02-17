package relay

import (
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
)

func (eng *Engine) SendGetBalance(target *Session, hook Callback) {
	hops := []byte{0, 1, 2, 3, 4, 5}
	s := make(Sessions, len(hops))
	s[2] = target
	se := eng.SelectHops(hops, s)
	var c Circuit
	copy(c[:], se)
	confID := nonce.NewID()
	o := GetBalance(target.ID, confID, se[5], c, eng.KeySet)
	log.D.Ln("sending out getbalance onion")
	res := eng.PostAcctOnion(o)
	eng.SendWithOneHook(c[0].AddrPort, res, hook)
}
