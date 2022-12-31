package wire

import (
	"github.com/Indra-Labs/indra/pkg/key/address"
	"github.com/Indra-Labs/indra/pkg/key/prv"
	"github.com/Indra-Labs/indra/pkg/key/pub"
	"github.com/Indra-Labs/indra/pkg/key/signer"
	"github.com/Indra-Labs/indra/pkg/node"
	"github.com/Indra-Labs/indra/pkg/nonce"
	"github.com/Indra-Labs/indra/pkg/slice"
)

// Ping is a message which checks the liveness of relays by ensuring they are
// correctly relaying messages. Pending pings are stored in a table with the
// last hop as the key to narrow the number of elements to search through to
// find the matching cipher and reveal the contained ID inside it.
//
// The pending ping records keep the identifiers of the three nodes that were in
// a ping onion and when the Confirmation is correctly received these nodes get
// an increment of their liveness score. By using this scheme, when nodes are
// offline their scores will fall to zero after a time whereas live nodes will
// have steadily increasing scores from successful pings.
func Ping(id nonce.ID, client *node.Node, hop [3]*node.Node,
	set *signer.KeySet) OnionSkins {

	n := GenPingNonces()
	return OnionSkins{}.
		Forward(hop[0].AddrPort).
		OnionSkin(address.FromPub(hop[0].HeaderPub), set.Next(), n[0]).
		Forward(hop[1].AddrPort).
		OnionSkin(address.FromPub(hop[1].HeaderPub), set.Next(), n[1]).
		Forward(hop[2].AddrPort).
		OnionSkin(address.FromPub(hop[2].HeaderPub), set.Next(), n[2]).
		Forward(client.AddrPort).
		OnionSkin(address.FromPub(client.HeaderPub), set.Next(), n[3]).
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
// the Reply relay.
//
// This message's last layer is a Confirmation, which allows the client to know
// that the key was successfully delivered to the Reply relays that will be
// used in the Purchase.
func SendKeys(id nonce.ID, hdr, pld *pub.Key,
	client *node.Node, hop [5]*node.Node, set *signer.KeySet) OnionSkins {

	n0 := Gen3Nonces()
	n1 := Gen3Nonces()
	return OnionSkins{}.
		Forward(hop[0].AddrPort).
		OnionSkin(address.FromPub(hop[0].HeaderPub), set.Next(), n0[0]).
		Forward(hop[1].AddrPort).
		OnionSkin(address.FromPub(hop[1].HeaderPub), set.Next(), n0[1]).
		Forward(hop[2].AddrPort).
		OnionSkin(address.FromPub(hop[2].HeaderPub), set.Next(), n0[2]).
		Cipher(hdr, pld).
		Forward(hop[3].AddrPort).
		OnionSkin(address.FromPub(hop[3].HeaderPub), set.Next(), n1[0]).
		Forward(hop[4].AddrPort).
		OnionSkin(address.FromPub(hop[4].HeaderPub), set.Next(), n1[1]).
		Forward(client.AddrPort).
		OnionSkin(address.FromPub(client.HeaderPub), set.Next(), n1[2]).
		Confirmation(id)
}

// SendPurchase delivers a request for keys for a relaying session with a given
// router (in this case, hop 2). It is almost identical to an Exit except the
// payload is always just a 64-bit unsigned integer.
//
// The response, which will be two public keys that identify the session and
// form the basis of the cloaked "To" keys, is encrypted with the given layers,
// the ciphers are already given in reverse order, so they are decoded in given
// order to create the correct payload encryption to match the PayloadPub
// combined with the header's given public From key.
//
// The header remains a constant size and each node in the Reply trims off
// their section at the top, moves the next layer header to the top and pads the
// remainder with noise, so it always looks like the first hop,
// indistinguishable.
func SendPurchase(nBytes uint64, client *node.Node,
	hop [5]*node.Node, set *signer.KeySet) OnionSkins {

	var prvs [3]*prv.Key
	for i := range prvs {
		prvs[i] = set.Next()
	}
	n0, n1 := Gen3Nonces(), Gen3Nonces()
	var pubs [3]*pub.Key
	pubs[0] = client.PayloadPub
	pubs[1] = hop[4].PayloadPub
	pubs[2] = hop[3].PayloadPub
	return OnionSkins{}.
		Forward(hop[0].AddrPort).
		OnionSkin(address.FromPub(hop[0].HeaderPub), set.Next(), n0[0]).
		Forward(hop[1].AddrPort).
		OnionSkin(address.FromPub(hop[1].HeaderPub), set.Next(), n0[1]).
		Forward(hop[2].AddrPort).
		OnionSkin(address.FromPub(hop[2].HeaderPub), set.Next(), n0[2]).
		Purchase(nBytes, prvs, pubs, n1).
		Reply(hop[3].AddrPort).
		OnionSkin(address.FromPub(hop[3].HeaderPub), prvs[0], n1[0]).
		Reply(hop[4].AddrPort).
		OnionSkin(address.FromPub(hop[4].HeaderPub), prvs[1], n1[1]).
		Reply(client.AddrPort).
		OnionSkin(address.FromPub(client.HeaderPub), prvs[2], n1[2])
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
// The header remains a constant size and each node in the Reply trims off
// their section at the top, moves the next layer header to the top and pads the
// remainder with noise, so it always looks like the first hop.
func SendExit(payload slice.Bytes, port uint16, client *node.Node,
	hop [5]*node.Node, set *signer.KeySet) OnionSkins {

	var prvs [3]*prv.Key
	for i := range prvs {
		prvs[i] = set.Next()
	}
	n0, n1 := Gen3Nonces(), Gen3Nonces()
	var pubs [3]*pub.Key
	pubs[0] = client.PayloadPub
	pubs[1] = hop[4].PayloadPub
	pubs[2] = hop[3].PayloadPub
	return OnionSkins{}.
		Forward(hop[0].AddrPort).
		OnionSkin(address.FromPub(hop[0].HeaderPub), set.Next(), n0[0]).
		Forward(hop[1].AddrPort).
		OnionSkin(address.FromPub(hop[1].HeaderPub), set.Next(), n0[1]).
		Forward(hop[2].AddrPort).
		OnionSkin(address.FromPub(hop[2].HeaderPub), set.Next(), n0[2]).
		Exit(port, prvs, pubs, payload).
		Reply(hop[3].AddrPort).
		OnionSkin(address.FromPub(hop[3].HeaderPub), prvs[0], n1[0]).
		Reply(hop[4].AddrPort).
		OnionSkin(address.FromPub(hop[4].HeaderPub), prvs[1], n1[1]).
		Reply(client.AddrPort).
		OnionSkin(address.FromPub(client.HeaderPub), prvs[2], n1[2])
}
