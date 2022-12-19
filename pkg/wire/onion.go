package wire

import (
	"github.com/Indra-Labs/indra/pkg/key/address"
	"github.com/Indra-Labs/indra/pkg/key/signer"
	"github.com/Indra-Labs/indra/pkg/node"
)

func Ping(nodes [3]node.Node, set signer.KeySet) Onion {
	return OnionSkins{}.
		Message(address.FromPubKey(nodes[0].Forward), set.Next()).
		Forward(nodes[0].IP).
		Message(address.FromPubKey(nodes[1].Forward), set.Next()).
		Forward(nodes[1].IP).
		Message(address.FromPubKey(nodes[2].Forward), set.Next()).
		Forward(nodes[2].IP).
		Assemble()
}
