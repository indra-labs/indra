package relay

import (
	"time"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/onion"
	"git-indra.lan/indra-labs/indra/pkg/traffic"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

func (eng *Engine) SendExit(port uint16, message slice.Bytes, id nonce.ID,
	target *traffic.Session, hook func(id nonce.ID, b slice.Bytes),
	timeout time.Duration) {
	
	hops := []byte{0, 1, 2, 3, 4, 5}
	s := make(traffic.Sessions, len(hops))
	s[2] = target
	se := eng.SelectHops(hops, s)
	var c traffic.Circuit
	copy(c[:], se)
	o := onion.SendExit(port, message, id, se[len(se)-1], c, eng.KeySet)
	log.D.Ln("sending out exit onion")
	res := eng.PostAcctOnion(o)
	eng.SendWithOneHook(c[0].AddrPort, res, hook)
}

func (eng *Engine) MakeExit(port uint16, message slice.Bytes, id nonce.ID,
	target *traffic.Session) (c traffic.Circuit,
	o onion.Skins) {
	
	hops := []byte{0, 1, 2, 3, 4, 5}
	s := make(traffic.Sessions, len(hops))
	s[2] = target
	se := eng.SelectHops(hops, s)
	copy(c[:], se)
	o = onion.SendExit(port, message, id, se[len(se)-1], c, eng.KeySet)
	return
}

func (eng *Engine) SendExitNew(c traffic.Circuit,
	o onion.Skins, hook func(id nonce.ID, b slice.Bytes),
	timeout time.Duration) {
	
	log.D.Ln("sending out exit onion")
	res := eng.PostAcctOnion(o)
	eng.SendWithOneHook(c[0].AddrPort, res, hook)
}
