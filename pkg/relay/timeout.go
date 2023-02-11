package relay

import (
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/traffic"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

// HandleTimeout is called automatically after an expected amount of time.
func (pr *PendingResponse) HandleTimeout(eng *Engine) func() {
	return func() {
		log.D.Ln("response timeout")
		var c traffic.Circuit
		for i := range pr.Billable {
			c[i] = eng.FindSession(pr.Billable[i])
		}
		log.D.Ln(c)
		nodes := make([]*traffic.Node, len(c))
		for i := range c {
			nodes[i] = c[i].Node
		}
		nodeIDs := make([]nonce.ID, len(c))
		for i := range nodes {
			nodeIDs[i] = nodes[i].ID
		}
		hooks := make([]func(id nonce.ID, b slice.Bytes), len(c))
		for i := range hooks {
			_ = i
		}
	}
}
