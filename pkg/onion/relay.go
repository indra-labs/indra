package onion

import (
	"net"

	"github.com/Indra-Labs/indra/pkg/node"
)

type Relay struct {
	net.IP
	node.Nodes
}
