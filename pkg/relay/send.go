package relay

import (
	"net/netip"
	"runtime"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

// Send a message to a peer via their AddrPort.
func (eng *Engine) Send(addr *netip.AddrPort, b slice.Bytes) {
	// first search if we already have the node available with connection
	// open.
	as := addr.String()
	eng.ForEachNode(func(n *Node) bool {
		if as == n.AddrPort.String() {
			_, f, l, _ := runtime.Caller(1)
			log.T.F("%s sending message to %v %s:%d",
				eng.GetLocalNode().AddrPort.String(), addr, f, l)
			n.Transport.Send(b)
			return true
		}
		return false
	})
}

// SendWithOneHook is used for onions with only one confirmation hook. Usually
// as returned from PostAcctOnion this is the last, confirmation or response
// layer in an onion.Skins.
func (eng *Engine) SendWithOneHook(ap *netip.AddrPort, res SendData,
	responseHook Callback) {
	
	if responseHook == nil {
		responseHook = func(_ nonce.ID, _ slice.Bytes) {
			log.D.Ln("nil response hook")
		}
	}
	eng.PendingResponses.Add(res.last, len(res.b), res.sessions, res.billable,
		res.ret, res.port, responseHook, res.postAcct)
	log.T.Ln("sending out onion")
	eng.Send(ap, res.b)
}
