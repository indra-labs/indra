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
//
// The first hop (0) is the destination of the first layer, 1 is second, 2 is
// the return relay, 3 is the first return, 4 is the second return, and client
// is the client.
func SendReturn(idCipher sha256.Hash, id nonce.ID, hdr, pld *prv.Key,
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
