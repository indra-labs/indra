package relay

import (
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/onion"
	"git-indra.lan/indra-labs/indra/pkg/traffic"
)

func (eng *Engine) SendGetBalance(target *traffic.Session, hook Callback) {
	hops := []byte{0, 1, 2, 3, 4, 5}
	s := make(traffic.Sessions, len(hops))
	s[2] = target
	se := eng.SelectHops(hops, s)
	var c traffic.Circuit
	copy(c[:], se)
	confID := nonce.NewID()
	o := onion.GetBalance(se[2].ID, confID, se[5], c, eng.KeySet)
	log.D.Ln("sending out exit onion")
	eng.SendOnion(c[0].AddrPort, o, hook, 0)
}
