package client

import (
	"math/rand"

	"github.com/indra-labs/indra/pkg/node"
	"github.com/indra-labs/indra/pkg/slice"
)

// SimpleSelector is a pure random shuffle selection algorithm for producing
// a list of nodes for an onion.
//
// This function will return nil if there isn't enough
func SimpleSelector(n node.Nodes, exit *node.Node,
	count int) (selected node.Nodes) {

	// For the purposes of this simple selector algorithm we require unique
	// nodes for each hop.
	if len(n) < count+1 {
		log.E.F("not enough nodes, have %d, need %d", len(n), count+1)
		return
	}
	// Remove the exit from the list of options.
	var nCandidates node.Nodes
	if exit != nil {
		for i := range n {
			if n[i].ID != exit.ID {
				nCandidates = append(nCandidates, n[i])
			}
		}
	}
	// Shuffle the list we made
	rand.Seed(slice.GetCryptoRandSeed())
	rand.Shuffle(len(nCandidates), func(i, j int) {
		nCandidates[i], nCandidates[j] = nCandidates[j], nCandidates[i]
	})
	return nCandidates[:count]
}
