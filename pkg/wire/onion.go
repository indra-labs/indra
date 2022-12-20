package wire

import (
	"github.com/Indra-Labs/indra/pkg/key/address"
	"github.com/Indra-Labs/indra/pkg/key/prv"
	"github.com/Indra-Labs/indra/pkg/key/signer"
	"github.com/Indra-Labs/indra/pkg/node"
	"github.com/Indra-Labs/indra/pkg/nonce"
	"github.com/Indra-Labs/indra/pkg/sha256"
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
func Ping(ciph sha256.Hash, id nonce.ID, nodes [3]node.Node,
	set signer.KeySet) Onion {

	return OnionSkins{}.
		Message(address.FromPubKey(nodes[0].Forward), set.Next()).
		Forward(nodes[0].IP).
		Message(address.FromPubKey(nodes[1].Forward), set.Next()).
		Forward(nodes[1].IP).
		Message(address.FromPubKey(nodes[2].Forward), set.Next()).
		Forward(nodes[2].IP).
		Confirmation(ciph, id).
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
func SendReturn(id nonce.ID, ciph sha256.Hash, hdr, pld *prv.Key,
	nodes [5]node.Node, set signer.KeySet) Onion {

	return OnionSkins{}.
		Message(address.FromPubKey(nodes[0].Forward), set.Next()).
		Forward(nodes[0].IP).
		Message(address.FromPubKey(nodes[1].Forward), set.Next()).
		Forward(nodes[1].IP).
		Message(address.FromPubKey(nodes[2].Forward), set.Next()).
		Cipher(hdr, pld).
		Message(address.FromPubKey(nodes[3].Forward), set.Next()).
		Forward(nodes[1].IP).
		Message(address.FromPubKey(nodes[4].Forward), set.Next()).
		Forward(nodes[2].IP).
		Confirmation(ciph, id).
		Assemble()
}
