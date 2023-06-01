package headers

import (
	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/engine/sessions"
	"github.com/indra-labs/indra/pkg/onions/exit"
)

type Headers struct {
	Forward, Return *exit.Routing
	ReturnPubs      crypto.Pubs
}

func GetHeaders(alice, bob *sessions.Data, c sessions.Circuit,
	ks *crypto.KeySet) (h *Headers) {

	fwKeys := ks.Next3()
	rtKeys := ks.Next3()
	n := crypto.GenNonces(6)
	var rtNonces, fwNonces [3]nonce.IV
	copy(fwNonces[:], n[:3])
	copy(rtNonces[:], n[3:])
	var fwSessions, rtSessions [3]*sessions.Data
	copy(fwSessions[:], c[:2])
	fwSessions[2] = bob
	copy(rtSessions[:], c[3:])
	rtSessions[2] = alice
	var returnPubs crypto.Pubs
	returnPubs[0] = c[3].Payload.Pub
	returnPubs[1] = c[4].Payload.Pub
	returnPubs[2] = alice.Payload.Pub
	h = &Headers{
		Forward: &exit.Routing{
			Sessions: fwSessions,
			Keys:     fwKeys,
			Nonces:   fwNonces,
		},
		Return: &exit.Routing{
			Sessions: rtSessions,
			Keys:     rtKeys,
			Nonces:   rtNonces,
		},
		ReturnPubs: returnPubs,
	}
	return
}

func (h *Headers) ExitPoint() *exit.ExitPoint {
	return &exit.ExitPoint{
		Routing:    h.Return,
		ReturnPubs: h.ReturnPubs,
	}
}
