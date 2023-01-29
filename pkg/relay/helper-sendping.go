package relay

import (
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/onion"
	"git-indra.lan/indra-labs/indra/pkg/traffic"
)

func (en *Engine) SendPing(c traffic.Circuit, conf Callback) {

	hops := []byte{0, 1, 2, 3, 4, 5}
	s := make(traffic.Sessions, len(hops))
	copy(s, c[:])
	se := en.Select(hops, s)
	copy(c[:], se)
	confID := nonce.NewID()
	o := onion.Ping(confID, se[len(se)-1], c, en.KeySet)
	en.SendOnion(c[0].AddrPort, o, conf)
}
