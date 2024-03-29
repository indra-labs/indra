package engine

import (
	"git.indra-labs.org/dev/ind/pkg/codec"
	"git.indra-labs.org/dev/ind/pkg/codec/onion/exit"
	"git.indra-labs.org/dev/ind/pkg/codec/ont"
	"git.indra-labs.org/dev/ind/pkg/crypto"
	"git.indra-labs.org/dev/ind/pkg/engine/sessions"
	"git.indra-labs.org/dev/ind/pkg/hidden"
)

// MakeReplyHeader constructs a reply header for hidden service messages.
func MakeReplyHeader(ng *Engine) (returnHeader *hidden.ReplyHeader) {
	n := crypto.GenNonces(3)
	rvKeys := ng.KeySet.Next3()
	hops := []byte{3, 4, 5}
	s := make(sessions.Sessions, len(hops))
	ng.Mgr().SelectHops(hops, s, "make message reply header")
	rt := &exit.Routing{
		Sessions: [3]*sessions.Data{s[0], s[1], s[2]},
		Keys:     crypto.Privs{rvKeys[0], rvKeys[1], rvKeys[2]},
		Nonces:   crypto.Nonces{n[0], n[1], n[2]},
	}
	rh := Skins{}.RoutingHeader(rt, ng.Mgr().Protocols)
	rHdr := codec.Encode(ont.Assemble(rh))
	rHdr.SetCursor(0)
	ep := exit.ExitPoint{
		Routing: rt,
		ReturnPubs: crypto.Pubs{
			crypto.DerivePub(s[0].Payload.Prv),
			crypto.DerivePub(s[1].Payload.Prv),
			crypto.DerivePub(s[2].Payload.Prv),
		},
	}
	returnHeader = &hidden.ReplyHeader{
		RoutingHeaderBytes: hidden.GetRoutingHeaderFromCursor(rHdr),
		Ciphers:            crypto.GenCiphers(ep.Routing.Keys, ep.ReturnPubs),
		Nonces:             ep.Routing.Nonces,
	}
	return
}
