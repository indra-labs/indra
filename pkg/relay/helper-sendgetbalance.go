package relay

import (
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/onion"
	"git-indra.lan/indra-labs/indra/pkg/traffic"
)

func (eng *Engine) SendGetBalance(s *traffic.Session, conf Callback) {
	var c traffic.Circuit
	var returns [3]*traffic.Session
	hops := make([]byte, 0)
	if s.Hop == 0 || s.Hop == 4 {
		hops = append(hops, s.Hop)
		c[s.Hop] = s
		hops = append(hops, 5)
		se := make(traffic.Sessions, len(hops))
		ss := eng.SessionManager.SelectHops(hops, se)
		returns[2] = ss[1]
		confID := nonce.NewID()
		o := onion.GetBalance(c, int(s.Hop), returns, eng.KeySet, confID)
		eng.SendOnion(c[s.Hop].AddrPort, o, conf, 0)
		return
	}
	var cur byte
	for i := 0; i < int(s.Hop); i++ {
		hops = append(hops, cur)
		cur++
	}
	hops = append(hops, s.Hop)
	for i := 3; i < 6; i++ {
		hops = append(hops, byte(i))
	}
	se := make(traffic.Sessions, len(hops))
	se[s.Hop] = s
	ss := eng.SessionManager.SelectHops(hops, se)
	// Construct the circuit parameter.
	for i := range ss {
		if i > int(s.Hop) {
			break
		}
		c[i] = ss[i]
	}
	lastIndex := len(hops) - 3
	for i := range returns {
		returns[i] = ss[lastIndex+i]
	}
	confID := nonce.NewID()
	o := onion.GetBalance(c, int(s.Hop), returns, eng.KeySet, confID)
	eng.SendOnion(c[0].AddrPort, o, conf, 0)
}
