package engine

import (
	"git-indra.lan/indra-labs/indra/pkg/crypto"
	"git-indra.lan/indra-labs/indra/pkg/engine/onions"
	"git-indra.lan/indra-labs/indra/pkg/engine/sessions"
)

func MakeReplyHeader(ng *Engine) (returnHeader *onions.ReplyHeader) {
	n := crypto.GenNonces(3)
	rvKeys := ng.KeySet.Next3()
	hops := []byte{3, 4, 5}
	s := make(sessions.Sessions, len(hops))
	ng.Manager.SelectHops(hops, s, "make message reply header")
	rt := &onions.Routing{
		Sessions: [3]*sessions.Data{s[0], s[1], s[2]},
		Keys:     crypto.Privs{rvKeys[0], rvKeys[1], rvKeys[2]},
		Nonces:   crypto.Nonces{n[0], n[1], n[2]},
	}
	rh := onions.Skins{}.RoutingHeader(rt)
	rHdr := onions.Encode(rh.Assemble())
	rHdr.SetCursor(0)
	ep := onions.ExitPoint{
		Routing: rt,
		ReturnPubs: crypto.Pubs{
			crypto.DerivePub(s[0].Payload.Prv),
			crypto.DerivePub(s[1].Payload.Prv),
			crypto.DerivePub(s[2].Payload.Prv),
		},
	}
	returnHeader = &onions.ReplyHeader{
		RoutingHeaderBytes: onions.GetRoutingHeaderFromCursor(rHdr),
		Ciphers:            crypto.GenCiphers(ep.Routing.Keys, ep.ReturnPubs),
		Nonces:             ep.Routing.Nonces,
	}
	return
}
