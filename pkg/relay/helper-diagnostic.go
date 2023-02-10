package relay

import (
	"git-indra.lan/indra-labs/indra/pkg/onion"
	"git-indra.lan/indra-labs/indra/pkg/traffic"
)

func (eng *Engine) SendDiagnostic(c *traffic.Circuit, hook Callback) {
	hops := []byte{
		0, 3, 4, 5,
		1, 3, 4, 5,
		2, 3, 4, 5,
		3, 3, 4, 5,
		4, 3, 4, 5,
	}
	s := make(traffic.Sessions, len(hops))
	for i := 0; i < 5; i++ {
		s[i*4] = c[i]
	}
	se := eng.SelectHops(hops, s)
	o := onion.Diagnostic(se, eng.KeySet)
	log.D.Ln("sending out diagnostic onion")
	log.T.S(o)
	eng.SendOnion(c[0].AddrPort, o, hook, 0)
}
