package client

import (
	"github.com/indra-labs/indra/pkg/crypto/sha256"
	"github.com/indra-labs/indra/pkg/onion/layers/balance"
	"github.com/indra-labs/indra/pkg/traffic"
	"github.com/indra-labs/indra/pkg/types"
	"github.com/indra-labs/indra/pkg/util/slice"
	"github.com/indra-labs/lnd/lnd/lnwire"
)

func (cl *Client) balance(on *balance.Layer,
	b slice.Bytes, c *slice.Cursor, prev types.Onion) {

	log.T.S(on.ConfID)

	cl.IterateSessions(func(s *traffic.Session) bool {
		if s.ID == on.ID {
			log.D.F("received balance %x for session %x",
				on.MilliSatoshi, on.ID)
			// todo: check for close match client's running estimate
			//  based on the outbound packet send volume.
			s.Remaining = on.MilliSatoshi
			return true
		}
		return false
	})
	pending := cl.PendingResponses.Find(sha256.Single(on.ID[:]))
	if pending != nil {
		for i := range pending.Billable {
			s := cl.FindSession(pending.Billable[i])
			if s != nil {
				log.D.Ln(cl.AddrPort.String(), "post acct")
				if i == 0 {
					cl.DecSession(s.ID,
						s.RelayRate*lnwire.MilliSatoshi(len(b)/2)/1024/1024)
				} else {
					cl.DecSession(s.ID,
						s.RelayRate*lnwire.MilliSatoshi(len(b))/1024/1024)
				}
			}
		}
		cl.PendingResponses.Delete(pending.Hash)
	}
	cl.Confirms.Confirm(on.ConfID)
}
