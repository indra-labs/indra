package relay

import (
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/onion"
	"git-indra.lan/indra-labs/indra/pkg/traffic"
)

func (eng *Engine) SendPing(c traffic.Circuit, conf Callback) {
	
	hops := []byte{0, 1, 2, 3, 4, 5}
	s := make(traffic.Sessions, len(hops))
	copy(s, c[:])
	se := eng.SelectHops(hops, s)
	copy(c[:], se)
	confID := nonce.NewID()
	o := onion.Ping(confID, se[len(se)-1], c, eng.KeySet)
	eng.SendOnion(c[0].AddrPort, o, conf, 0)
}
