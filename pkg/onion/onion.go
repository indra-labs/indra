// Package onion provides a set of functions to manage creating onion layered
// encryption for use with multi-hop Circuit protocol.
package onion

import (
	"github.com/Indra-Labs/indra/pkg/node"
)

type Circuit struct {
	Hops [5]*node.Node
}
