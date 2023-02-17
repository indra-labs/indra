package relay

import (
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
)

func (eng *Engine) SendPing(c Circuit, hook Callback) {
	hops := []byte{0, 1, 2, 3, 4, 5}
	s := make(Sessions, len(hops))
	copy(s, c[:])
	se := eng.SelectHops(hops, s)
	copy(c[:], se)
	confID := nonce.NewID()
	o := Ping(confID, se[len(se)-1], c, eng.KeySet)
	res := eng.PostAcctOnion(o)
	eng.SendWithOneHook(c[0].AddrPort, res, hook)
}
