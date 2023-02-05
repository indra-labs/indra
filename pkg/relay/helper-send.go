package relay

import (
	"net/netip"
	"runtime"
	
	"git-indra.lan/indra-labs/indra/pkg/traffic"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

// Send a message to a peer via their AddrPort.
func (eng *Engine) Send(addr *netip.AddrPort, b slice.Bytes) {
	// first search if we already have the node available with connection
	// open.
	as := addr.String()
	eng.ForEachNode(func(n *traffic.Node) bool {
		if as == n.AddrPort.String() {
			n.Transport.Send(b)
			_, f, l, _ := runtime.Caller(1)
			log.T.F("%s sending message to %v %s:%d",
				eng.GetLocalNode().AddrPort.String(), addr, f, l)
			return true
		}
		return false
	})
}
