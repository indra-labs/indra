package indra

import (
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/onion"
	"github.com/indra-labs/indra/pkg/traffic"
)

func (en *Engine) SendGetBalance(s *traffic.Session, conf func(cf nonce.ID)) {
	var c traffic.Circuit
	var returns [3]*traffic.Session
	hops := make([]byte, 0)
	if s.Hop == 0 || s.Hop == 4 {
		hops = append(hops, s.Hop)
		c[s.Hop] = s
		hops = append(hops, 5)
		se := make(traffic.Sessions, len(hops))
		ss := en.Payments.Select(hops, se)
		returns[2] = ss[1]
		confID := nonce.NewID()
		o := onion.GetBalance(c, int(s.Hop), returns, en.KeySet, confID)
		en.RegisterConfirmation(conf, confID)
		en.SendOnion(c[s.Hop].AddrPort, o, nil)
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
	ss := en.Payments.Select(hops, se)
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
	o := onion.GetBalance(c, int(s.Hop), returns, en.KeySet, confID)
	en.RegisterConfirmation(conf, confID)
	en.SendOnion(c[0].AddrPort, o, nil)
}
