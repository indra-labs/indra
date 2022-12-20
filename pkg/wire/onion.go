package wire

import (
	"github.com/Indra-Labs/indra/pkg/key/address"
	"github.com/Indra-Labs/indra/pkg/key/prv"
	"github.com/Indra-Labs/indra/pkg/key/signer"
	"github.com/Indra-Labs/indra/pkg/node"
	"github.com/Indra-Labs/indra/pkg/nonce"
	"github.com/Indra-Labs/indra/pkg/sha256"
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
func Ping(id nonce.ID, client node.Node, hop [3]node.Node,
	set signer.KeySet) Onion {

	return OnionSkins{}.
		Message(address.FromPubKey(hop[0].HeaderKey), set.Next()).
		Forward(hop[1].IP).
		Message(address.FromPubKey(hop[1].HeaderKey), set.Next()).
		Forward(hop[2].IP).
		Message(address.FromPubKey(hop[2].HeaderKey), set.Next()).
		Forward(client.IP).
		Message(address.FromPubKey(client.HeaderKey), set.Next()).
		Confirmation(id).
		Assemble()
}

// SendReturn provides a pair of private keys that will be used to generate the
// Purchase header bytes and to generate the ciphers provided in the Purchase
// message to encrypt the Session that is returned.
//
// The Header key, its cloaked public key counterpart used in the To field of
// the Purchase message preformed header bytes, but the Ciphers provided in the
// Purchase message, for encrypting the Session to be returned, uses the Payload
// key, along with the public key found in the encrypted layer of the header for
// the Return relay.
//
// This message's last layer is a Confirmation, which allows the client to know
// that the key was successfully delivered to the Return relays that will be
// used in the Purchase.
func SendReturn(id nonce.ID, hdr, pld *prv.Key,
	client node.Node, hop [5]node.Node, set signer.KeySet) Onion {

	return OnionSkins{}.
		Message(address.FromPubKey(hop[0].HeaderKey), set.Next()).
		Forward(hop[1].IP).
		Message(address.FromPubKey(hop[1].HeaderKey), set.Next()).
		Forward(hop[2].IP).
		Message(address.FromPubKey(hop[2].HeaderKey), set.Next()).
		Cipher(hdr, pld).
		Forward(hop[3].IP).
		Message(address.FromPubKey(hop[3].HeaderKey), set.Next()).
		Forward(hop[4].IP).
		Message(address.FromPubKey(hop[4].HeaderKey), set.Next()).
		Forward(client.IP).
		Message(address.FromPubKey(client.HeaderKey), set.Next()).
		Confirmation(id).
		Assemble()
}

// SendExit constructs a message containing an arbitrary payload to a node (3rd
// hop) with a set of 3 ciphers derived from the hidden PayloadKey of the return
// hops that are layered progressively after the Exit message.
//
// The Exit node forwards the packet it receives to the local port specified in
// the Exit message, and then uses the ciphers to encrypt the return with the
// three ciphers provided, which don't enable it to decrypt the header, only to
// encrypt the payload.
//
// TODO: we can create the ciphers based on hop 3, 4 and client Nodes.
func SendExit(payload slice.Bytes, port uint16, ciphers [3]sha256.Hash,
	client node.Node, hop [5]node.Node, set signer.KeySet) Onion {

	return OnionSkins{}.
		Message(address.FromPubKey(hop[0].HeaderKey), set.Next()).
		Forward(hop[1].IP).
		Message(address.FromPubKey(hop[1].HeaderKey), set.Next()).
		Forward(hop[2].IP).
		Message(address.FromPubKey(hop[2].HeaderKey), set.Next()).
		Exit(port, ciphers, payload).
		Return(hop[3].IP).
		Message(address.FromPubKey(hop[3].PayloadKey), set.Next()).
		Return(hop[4].IP).
		Message(address.FromPubKey(hop[4].PayloadKey), set.Next()).
		Return(client.IP).
		Message(address.FromPubKey(client.PayloadKey), set.Next()).
		Assemble()
}
