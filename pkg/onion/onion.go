package onion

import (
	"github.com/indra-labs/indra/pkg/crypto/key/prv"
	"github.com/indra-labs/indra/pkg/crypto/key/pub"
	"github.com/indra-labs/indra/pkg/crypto/key/signer"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/node"
	"github.com/indra-labs/indra/pkg/onion/layers/session"
	"github.com/indra-labs/indra/pkg/traffic"
	"github.com/indra-labs/indra/pkg/util/slice"
)

// Ping is a message which checks the liveness of relays by ensuring they are
// correctly relaying messages.
//
// The pending ping records keep the identifiers of the 5 nodes that were in
// a ping onion and when the Confirmation is correctly received these nodes get
// an increment of their liveness score. By using this scheme, when nodes are
// offline their scores will fall to zero after a time whereas live nodes will
// have steadily increasing scores from successful pings.
func Ping(id nonce.ID, client *traffic.Session, s traffic.Circuit,
	ks *signer.KeySet) Skins {

	n := GenPingNonces()
	return Skins{}.
		ForwardCrypt(s[0], ks.Next(), n[0]).
		ForwardCrypt(s[1], ks.Next(), n[1]).
		ForwardCrypt(s[2], ks.Next(), n[2]).
		ForwardCrypt(s[3], ks.Next(), n[3]).
		ForwardCrypt(s[4], ks.Next(), n[4]).
		ForwardCrypt(client, ks.Next(), n[5]).
		Confirmation(id)
}

// SendExit constructs a message containing an arbitrary payload to a node (3rd
// hop) with a set of 3 ciphers derived from the hidden PayloadPub of the return
// hops that are layered progressively after the Exit message.
//
// The Exit node forwards the packet it receives to the local port specified in
// the Exit message, and then uses the ciphers to encrypt the reply with the
// three ciphers provided, which don't enable it to decrypt the header, only to
// encrypt the payload.
//
// The response is encrypted with the given layers, the ciphers are already
// given in reverse order, so they are decoded in given order to create the
// correct payload encryption to match the PayloadPub combined with the header's
// given public From key.
//
// The header remains a constant size and each node in the Reverse trims off
// their section at the top, moves the next crypt header to the top and pads the
// remainder with noise, so it always looks like the first hop.
func SendExit(port uint16, payload slice.Bytes,
	client *traffic.Session, s traffic.Circuit, ks *signer.KeySet) Skins {

	var prvs [3]*prv.Key
	for i := range prvs {
		prvs[i] = ks.Next()
	}
	n := GenNonces(6)
	var returnNonces [3]nonce.IV
	copy(returnNonces[:], n[3:])
	var pubs [3]*pub.Key
	pubs[0] = s[3].PayloadPub
	pubs[1] = s[4].PayloadPub
	pubs[2] = client.PayloadPub
	return Skins{}.
		ReverseCrypt(s[0], ks.Next(), n[0], 3).
		ReverseCrypt(s[1], ks.Next(), n[1], 2).
		ReverseCrypt(s[2], ks.Next(), n[2], 1).
		Exit(port, prvs, pubs, returnNonces, payload).
		ReverseCrypt(s[3], prvs[0], n[3], 0).
		ReverseCrypt(s[4], prvs[1], n[4], 0).
		ReverseCrypt(client, prvs[2], n[5], 0)
}

// SendKeys provides a pair of private keys that will be used to generate the
// Purchase header bytes and to generate the ciphers provided in the Purchase
// message to encrypt the Session that is returned.
//
// The OnionSkin key, its cloaked public key counterpart used in the ToHeaderPub field of
// the Purchase message preformed header bytes, but the Ciphers provided in the
// Purchase message, for encrypting the Session to be returned, uses the Payload
// key, along with the public key found in the encrypted crypt of the header for
// the Reverse relay.
//
// This message's last crypt is a Confirmation, which allows the client to know
// that the keys were successfully delivered.
//
// This is the only onion that uses the node identity keys. The payment preimage
// hash must be available or the relay should not forward the remainder of the
// packet.
//
// If hdr/pld cipher keys are nil there must be a HeaderPub available on the
// session for the hop. This allows this function to send keys to any number of
// hops, but the very first SendKeys must have all in order to create the first
// set of sessions. This is by way of indicating to not use the IdentityPub but
// the HeaderPub instead. Not allowing free relay at all prevents spam attacks.
func SendKeys(id nonce.ID, s [5]*session.Layer,
	client *traffic.Session, hop node.Nodes, ks *signer.KeySet) Skins {

	n := GenNonces(6)
	sk := Skins{}
	for i := range s {
		sk.ForwardSession(hop[i], ks.Next(), n[i], s[i])
	}
	return sk.
		ForwardCrypt(client, ks.Next(), n[5]).
		Confirmation(id)
}

// GetBalance sends out a request in a similar way to SendExit except the node
// being queried can be any of the 5 and the return path is always a further two
// hops until the client.
//
// First and last hop sessions are just directly queried, and the rest the path
// goes to them and then through two return hops in the second last and last
// positions, because it cannot be concealed that the client is not a relay.
//
// The first and last hops already have a payload session key ready to use and
// all that is then sent back is a crypt with the query ID and session balance
// addressed to the Header key.
func GetBalance(s traffic.Circuit, target int, returns [3]*traffic.Session,
	ks *signer.KeySet, id nonce.ID) (o Skins) {

	if target == 0 || target == 4 {
		n := GenNonces(2)
		o = o.ForwardCrypt(s[target], ks.Next(), n[0]).
			DirectBalance(s[target].ID, id).
			Forward(returns[2].AddrPort)
		log.T.S(o)
		return
	}
	n := GenNonces(target + 1 + 3)
	var returnNonces [3]nonce.IV
	copy(returnNonces[:], n[len(n)-1-3:])
	var prvs [3]*prv.Key
	for i := range prvs {
		prvs[i] = ks.Next()
	}
	var pubs [3]*pub.Key
	for i := range returns {
		pubs[i] = returns[i].PayloadPub
	}
	for i := 0; i <= target; i++ {
		o = o.ForwardCrypt(s[i], ks.Next(), n[i])
	}
	o = o.GetBalance(s[target].ID, id, prvs, pubs, returnNonces)
	for i := range returns {
		o = o.ReverseCrypt(returns[i], prvs[i], n[i+target], 0)
	}
	return
}
