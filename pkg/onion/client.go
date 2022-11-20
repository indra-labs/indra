package onion

import (
	"github.com/Indra-Labs/indra/pkg/node"
)

type Client struct {
	node.Nodes
	Circuits
}
