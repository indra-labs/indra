package indra

import (
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/node"
	"git-indra.lan/indra-labs/indra/pkg/onion"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/session"
	"git-indra.lan/indra-labs/indra/pkg/payment"
	"git-indra.lan/indra-labs/indra/pkg/traffic"
)

func (en *Engine) SendKeys(sb []*SessionBuy, sess []*session.Layer,
	pmt []*payment.Payment, hook func(hops []*traffic.Session)) {

	if len(sb) != len(sess) || len(sess) != len(pmt) {
		log.E.Ln("all sendkeys parameters must be same length")
		return
	}
	var buys [][]*SessionBuy
	var s [][]*session.Layer
	var p [][]*payment.Payment
	n := len(sb) / 5
	nMod := len(sb) % 5
	if nMod != 0 {
		n++
	}
	for i := 0; i < n; i++ {
		buys = append(buys, sb[i*5:][:5])
		s = append(s, sess[i*5:][:5])
		p = append(p, pmt[i*5:][:5])
	}
	for bu := range buys {
		hops := []byte{0, 1, 2, 3, 4, 5}
		sessions := make(traffic.Sessions, len(hops))
		// Put the sessions in the middle if there is less than 5.
		for i := range s {
			sessions[i] = traffic.NewSession(nonce.NewID(),
				buys[bu][i].Peer, buys[bu][i].Amount,
				s[bu][i].Header, s[bu][i].Payload, byte(i))
		}
		// Fill the gaps.
		se := en.Select([]byte{5}, make(traffic.Sessions, 1))
		cnf := nonce.NewID()
		// Send the keys.
		var circuit node.Nodes
		for i := range buys[bu] {
			circuit = append(circuit, buys[bu][i].Node)
		}
		// Build the session layer parameter.
		var ss [5]*session.Layer
		for i := range s[bu] {
			ss[i] = s[bu][i]
		}
		// FIRE!
		sk := onion.SendKeys(cnf, ss, se[0],
			circuit, en.KeySet)
		en.RegisterConfirmation(func(cf nonce.ID) {
			log.T.F("confirmed sendkeys id %x", cf)
			var h []*traffic.Session
			for i := range circuit {
				if circuit[i] != nil {
					h = append(h, sessions[i])
				}
			}
			hook(h)
		}, cnf)
		log.T.F("sending out %d session keys", len(buys[bu]))
		en.SendOnion(circuit[0].AddrPort, sk, nil)
	}
}
