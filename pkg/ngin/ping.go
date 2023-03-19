package ngin

import (
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/signer"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
)

// Ping is a message which checks the liveness of relays by ensuring they are
// correctly relaying messages.
//
// The pending ping records keep the identifiers of the 5 nodes that were in
// a ping onion and when the Confirmation is correctly received these nodes get
// an increment of their liveness score. By using this scheme, when nodes are
// offline their scores will fall to zero after a time whereas live nodes will
// have steadily increasing scores from successful pings.
func Ping(id nonce.ID, client *SessionData, s Circuit,
	ks *signer.KeySet) Skins {
	
	n := GenPingNonces()
	return Skins{}.
		Crypt(s[0].HeaderPub, nil, ks.Next(), n[0], 0).
		ForwardCrypt(s[1], ks.Next(), n[1]).
		ForwardCrypt(s[2], ks.Next(), n[2]).
		ForwardCrypt(s[3], ks.Next(), n[3]).
		ForwardCrypt(s[4], ks.Next(), n[4]).
		ForwardCrypt(client, ks.Next(), n[5]).
		Confirmation(id, 0)
}

func (ng *Engine) SendPing(c Circuit, hook Callback) {
	hops := StandardCircuit()
	s := make(Sessions, len(hops))
	copy(s, c[:])
	se := ng.SelectHops(hops, s)
	copy(c[:], se)
	confID := nonce.NewID()
	o := Ping(confID, se[len(se)-1], c, ng.KeySet)
	res := ng.PostAcctOnion(o)
	ng.SendWithOneHook(c[0].AddrPort, res, hook, ng.PendingResponses)
}
