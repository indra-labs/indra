package onions

import (
	"git-indra.lan/indra-labs/indra/pkg/util/qu"
	"git-indra.lan/indra-labs/indra/pkg/util/splice"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto"
	"git-indra.lan/indra-labs/indra/pkg/engine/coding"
	"git-indra.lan/indra-labs/indra/pkg/engine/responses"
	"git-indra.lan/indra-labs/indra/pkg/engine/sess"
	"git-indra.lan/indra-labs/indra/pkg/engine/sessions"
)

// Onion are messages that can be layered over each other and have
// a set of processing instructions for the data in them, and, if relevant,
// how to account for them in sessions.
type Onion interface {
	coding.Codec
	Wrap(inner Onion)
	Handle(s *splice.Splice, p Onion, ni Ngin) (e error)
	Account(res *sess.Data, sm *sess.Manager, s *sessions.Data,
		last bool) (skip bool, sd *sessions.Data)
}

type Ngin interface {
	HandleMessage(s *splice.Splice, pr Onion)
	GetLoad() byte
	SetLoad(byte)
	Mgr() *sess.Manager
	Pending() *responses.Pending
	GetHidden() *Hidden
	KillSwitch() qu.C
	Keyset() *crypto.KeySet
}
