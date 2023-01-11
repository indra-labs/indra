package wire

import (
	"github.com/indra-labs/indra/pkg/key/prv"
	"github.com/indra-labs/indra/pkg/key/pub"
	"github.com/indra-labs/indra/pkg/key/signer"
	"github.com/indra-labs/indra/pkg/node"
	"github.com/indra-labs/indra/pkg/nonce"
	"github.com/indra-labs/indra/pkg/session"
	"github.com/indra-labs/indra/pkg/slice"
)

// Ping is a message which checks the liveness of relays by ensuring they are
// correctly relaying messages. Pending pings are stored in a table with the
// last hop as the key to narrow the number of elements to search through to
// find the matching cipher and reveal the contained ID inside it.
//
// The pending ping records keep the identifiers of the 5 nodes that were in
// a ping onion and when the Confirmation is correctly received these nodes get
// an increment of their liveness score. By using this scheme, when nodes are
// offline their scores will fall to zero after a time whereas live nodes will
// have steadily increasing scores from successful pings.
func Ping(id nonce.ID, s session.Sessions, ks *signer.KeySet) OnionSkins {
	if len(s) != 6 {
		log.E.F("Ping requires 6 sessions, received %d", len(s))
		return nil
	}
	n := GenPingNonces()
	return OnionSkins{}.
		Forward(s[0].AddrPort).
		OnionSkin(s[0].HeaderPub, ks.Next(), n[0]).
		Forward(s[1].AddrPort).
		OnionSkin(s[1].HeaderPub, ks.Next(), n[1]).
		Forward(s[2].AddrPort).
		OnionSkin(s[2].HeaderPub, ks.Next(), n[2]).
		Forward(s[3].AddrPort).
		OnionSkin(s[3].HeaderPub, ks.Next(), n[3]).
		Forward(s[4].AddrPort).
		OnionSkin(s[4].HeaderPub, ks.Next(), n[3]).
		Forward(s[5].AddrPort).
		OnionSkin(s[5].HeaderPub, ks.Next(), n[3]).
		Confirmation(id)
}

// SendKeys provides a pair of private keys that will be used to generate the
// Purchase header bytes and to generate the ciphers provided in the Purchase
// message to encrypt the Session that is returned.
//
// The OnionSkin key, its cloaked public key counterpart used in the To field of
// the Purchase message preformed header bytes, but the Ciphers provided in the
// Purchase message, for encrypting the Session to be returned, uses the Payload
// key, along with the public key found in the encrypted layer of the header for
// the Reverse relay.
//
// This message's last layer is a Confirmation, which allows the client to know
// that the keys were successfully delivered.
func SendKeys(id nonce.ID, hdr, pld []*prv.Key,
	client *node.Node, hop []*node.Node, set *signer.KeySet) OnionSkins {

	n := GenNonces(6)
	return OnionSkins{}.
		Forward(hop[0].AddrPort).
		OnionSkin(hop[0].IdentityPub, set.Next(), n[0]).
		Cipher(hdr[0], pld[0]).
		Forward(hop[1].AddrPort).
		OnionSkin(hop[1].IdentityPub, set.Next(), n[1]).
		Cipher(hdr[1], pld[1]).
		Forward(hop[2].AddrPort).
		OnionSkin(hop[2].IdentityPub, set.Next(), n[2]).
		Cipher(hdr[2], pld[2]).
		Forward(hop[3].AddrPort).
		OnionSkin(hop[3].IdentityPub, set.Next(), n[3]).
		Cipher(hdr[3], pld[3]).
		Forward(hop[4].AddrPort).
		OnionSkin(hop[4].IdentityPub, set.Next(), n[4]).
		Cipher(hdr[4], pld[4]).
		Forward(client.AddrPort).
		OnionSkin(client.IdentityPub, set.Next(), n[5]).
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
// their section at the top, moves the next layer header to the top and pads the
// remainder with noise, so it always looks like the first hop.
func SendExit(payload slice.Bytes, port uint16, client *node.Node,
	hop [5]*node.Node, sess [3]*session.Session, set *signer.KeySet) OnionSkins {

	var prvs [3]*prv.Key
	for i := range prvs {
		prvs[i] = set.Next()
	}
	n0, n1 := Gen3Nonces(), Gen3Nonces()
	var pubs [3]*pub.Key
	pubs[0] = sess[0].PayloadPub
	pubs[1] = sess[1].PayloadPub
	pubs[2] = sess[2].PayloadPub
	return OnionSkins{}.
		Forward(hop[0].AddrPort).
		OnionSkin(hop[0].IdentityPub, set.Next(), n0[0]).
		Forward(hop[1].AddrPort).
		OnionSkin(hop[1].IdentityPub, set.Next(), n0[1]).
		Forward(hop[2].AddrPort).
		OnionSkin(hop[2].IdentityPub, set.Next(), n0[2]).
		Exit(port, prvs, pubs, n1, payload).
		Reverse(hop[3].AddrPort).
		OnionSkin(sess[0].HeaderPub, prvs[0], n1[0]).
		Reverse(hop[4].AddrPort).
		OnionSkin(sess[1].HeaderPub, prvs[1], n1[1]).
		Reverse(client.AddrPort).
		OnionSkin(sess[2].HeaderPub, prvs[2], n1[2])
}
