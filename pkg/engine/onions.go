package engine

import (
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/util/octet"
)

type Skins []Onion

var nop = &Tmpl{}

func Encode(on Onion) (s *octet.Splice) {
	s = octet.New(on.Len())
	check(on.Encode(s))
	return
}

// Assemble inserts the slice of Layer s inside each other so the first then
// contains the second, second contains the third, and so on, and then returns
// the first onion, on which you can then call Encode and generate the wire
// message form of the onion.
func (o Skins) Assemble() (on Onion) {
	// First item is the outer crypt.
	on = o[0]
	// Iterate through the remaining layers.
	for _, oc := range o[1:] {
		on.Wrap(oc)
		// Next step we are inserting inside the one we just inserted.
		on = oc
	}
	// At the end, the first element contains references to every element
	// inside it.
	return o[0]
}

func (o Skins) ForwardCrypt(s *SessionData, k *prv.Key, n nonce.IV) Skins {
	return o.Forward(s.AddrPort).Crypt(s.HeaderPub, s.PayloadPub, k, n, 0)
}

func (o Skins) ReverseCrypt(s *SessionData, k *prv.Key, n nonce.IV,
	seq int) Skins {
	
	return o.Reverse(s.AddrPort).Crypt(s.HeaderPub, s.PayloadPub, k, n, seq)
}

func (o Skins) RoutingHeader(s [3]*SessionData, fl [3]*prv.Key,
	n [3]nonce.IV) Skins {
	return o.
		ReverseCrypt(s[0], fl[0], n[0], 3).
		ReverseCrypt(s[1], fl[1], n[1], 2).
		ReverseCrypt(s[2], fl[2], n[2], 1)
}

func (o Skins) ForwardSession(s *Node,
	k *prv.Key, n nonce.IV, sess *Session) Skins {
	
	return o.Forward(s.AddrPort).
		Crypt(s.IdentityPub, nil, k, n, 0).
		Session(sess)
}
