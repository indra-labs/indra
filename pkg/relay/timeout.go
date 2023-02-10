package relay

import "git-indra.lan/indra-labs/indra/pkg/traffic"

// HandleTimeout is called automatically after an expected amount of time.
func (pr *PendingResponse) HandleTimeout(eng *Engine) func() {
	return func() {
		log.D.Ln("response timeout")
		var c traffic.Circuit
		for i := range pr.Billable {
			c[i] = eng.FindSession(pr.Billable[i])
		}
		log.D.Ln(c)
	}
}
