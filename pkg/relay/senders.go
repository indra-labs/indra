package relay

import (
	"net/netip"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

// SendWithOneHook is used for onions with only one confirmation hook. Usually
// as returned from PostAcctOnion this is the last, confirmation or response
// layer in an onion.Skins.
func (eng *Engine) SendWithOneHook(ap *netip.AddrPort, res SendData, responseHook func(id nonce.ID, b slice.Bytes), ) {
	
	if responseHook == nil {
		responseHook = func(_ nonce.ID, _ slice.Bytes) {
			log.D.Ln("nil response hook")
		}
	}
	eng.PendingResponses.Add(res.last, len(res.b), res.sessions, res.billable, res.ret, res.port, responseHook, res.postAcct)
	log.T.Ln("sending out onion")
	eng.Send(ap, res.b)
}
