package wire

import (
	"github.com/Indra-Labs/indra/pkg/key/address"
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
