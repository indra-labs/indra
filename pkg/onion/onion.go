package onion

import (
	"github.com/Indra-Labs/indra/pkg/node"
)

type Circuit struct {
	Hops [5]*node.Node
}
