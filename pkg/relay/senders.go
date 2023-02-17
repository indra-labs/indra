package relay

import (
	"net/netip"
	"time"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

// SendWithOneHook is used for onions with only one confirmation hook. Usually
// as returned from PostAcctOnion this is the last, confirmation or response
// layer in an onion.Skins.
func (eng *Engine) SendWithOneHook(
	ap *netip.AddrPort,
	res SendData,
	timeout time.Duration,
	responseHook func(id nonce.ID, b slice.Bytes),
) {
	
	if responseHook == nil {
		responseHook = func(_ nonce.ID, _ slice.Bytes) {
			log.D.Ln("nil response hook")
		}
	}
	eng.PendingResponses.Add(res.last, len(res.b), res.sessions, res.billable,
		res.ret, res.port, responseHook, res.postAcct, timeout, eng)
	log.T.Ln("sending out onion")
	eng.Send(ap, res.b)
}

// SendWithHooks allows multiple different hooks to be created that are
// identified by the IDs given to them. This is used for the diagnostic probe
// onion, which has a confirmation message back from all 5 hops of a circuit.
func (eng *Engine) SendWithHooks(
	ap *netip.AddrPort,
	res SendData,
	timeout time.Duration,
	ids []nonce.ID,
	responseHooks []Callback,
) {
	if len(responseHooks) != len(ids) {
		panic("programmer error, " +
			"must be equal number of response hooks and ids")
	}
	for i, responseHook := range responseHooks {
		if responseHook == nil {
			responseHook = func(_ nonce.ID, _ slice.Bytes) {
				log.D.Ln("nil response hook")
			}
		}
		eng.PendingResponses.Add(ids[i], len(res.b), res.sessions, res.billable,
			res.ret, res.port, responseHook, res.postAcct, timeout, eng)
	}
	log.T.Ln("sending out onion")
	eng.Send(ap, res.b)
}
