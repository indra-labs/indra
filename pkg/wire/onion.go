package wire

import (
	"github.com/indra-labs/indra/pkg/key/prv"
	"github.com/indra-labs/indra/pkg/key/pub"
	"github.com/indra-labs/indra/pkg/key/signer"
	"github.com/indra-labs/indra/pkg/node"
	"github.com/indra-labs/indra/pkg/nonce"
	"github.com/indra-labs/indra/pkg/slice"
)

// Ping is a message which checks the liveness of relays by ensuring they are
// correctly relaying messages.
//
// The pending ping records keep the identifiers of the 5 nodes that were in
// a ping onion and when the Confirmation is correctly received these nodes get
// an increment of their liveness score. By using this scheme, when nodes are
// offline their scores will fall to zero after a time whereas live nodes will
// have steadily increasing scores from successful pings.
func Ping(id nonce.ID, client *node.Session, s node.Circuit, ks *signer.KeySet) OnionSkins {
	n := GenPingNonces()
	return OnionSkins{}.
		ForwardLayer(s[0], ks.Next(), n[0]).
		ForwardLayer(s[1], ks.Next(), n[1]).
		ForwardLayer(s[2], ks.Next(), n[2]).
		ForwardLayer(s[3], ks.Next(), n[3]).
		ForwardLayer(s[4], ks.Next(), n[4]).
		ForwardLayer(client, ks.Next(), n[5]).
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
func SendExit(payload slice.Bytes, port uint16,
	client *node.Session, s node.Circuit, ks *signer.KeySet) OnionSkins {

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
	return OnionSkins{}.
		ForwardLayer(s[0], ks.Next(), n[0]).
		ForwardLayer(s[1], ks.Next(), n[1]).
		ForwardLayer(s[2], ks.Next(), n[2]).
		Exit(port, prvs, pubs, returnNonces, payload).
		ReverseLayer(s[3], prvs[0], n[3]).
		ReverseLayer(s[4], prvs[1], n[4]).
		ReverseLayer(client, prvs[2], n[5])
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
func SendKeys(id nonce.ID, hdr, pld [5]*prv.Key,
	client *node.Session, hop node.Circuit, ks *signer.KeySet) OnionSkins {

	n := GenNonces(6)
	return OnionSkins{}.
		ForwardSession(hop[0], ks.Next(), n[0], hdr[0], pld[0]).
		ForwardSession(hop[1], ks.Next(), n[1], hdr[1], pld[1]).
		ForwardSession(hop[2], ks.Next(), n[2], hdr[2], pld[2]).
		ForwardSession(hop[3], ks.Next(), n[3], hdr[3], pld[3]).
		ForwardSession(hop[4], ks.Next(), n[4], hdr[4], pld[4]).
		ForwardLayer(client, ks.Next(), n[5]).
		Confirmation(id)
}

// GetBalance sends out a request in a similar way to SendExit except the node
// being queried can be any of the 5 and the return path is always a further two
// hops until the client.
//
// The third returns Session should be the client's return session, index 0.
func GetBalance(s node.Circuit, target int,
	returns [3]*node.Session, ks *signer.KeySet) (o OnionSkins) {

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

	for i := 0; i < target; i++ {
		o = o.ForwardLayer(s[i], ks.Next(), n[i])
	}
	reqNonce := nonce.NewID()
	o = o.GetBalance(reqNonce, prvs, pubs, returnNonces)
	for i := range returns {
		o = o.ReverseLayer(returns[i], prvs[i], n[i+target])
	}
	o = o.Confirmation(reqNonce)
	return
}
