package client

import (
	"testing"

	"github.com/Indra-Labs/indra/pkg/node"
	"github.com/Indra-Labs/indra/pkg/nonce"
	"github.com/Indra-Labs/indra/pkg/transport"
)

func TestClient(t *testing.T) {
	var nodes node.Nodes
	var ids []nonce.ID
	nNodes := 10
	for i := 0; i < nNodes; i++ {
		n, id := node.New(nil, transport.NewSim(0))
		nodes = append(nodes, n)
		ids = append(ids, id)
	}
	cl := New(transport.NewSim(0))
	cl.Nodes = nodes
	
}
