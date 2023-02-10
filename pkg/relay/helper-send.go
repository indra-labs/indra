package relay

import (
	"net/netip"

	"github.com/davecgh/go-spew/spew"

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
			log.T.C(func() string {
				return eng.GetLocalNodeAddress().String() +
					" sending to " +
					addr.String() +
					"\n" +
					spew.Sdump(b.ToBytes())
			})
			n.Transport.Send(b)
			return true
		}
		return false
	})
	// If we got to here none of the addresses matched, and we need to
	// establish a new peer connection to them, if we know of them (this
	// would usually be the reason this happens).
	log.T.Ln("could not find peer with address", addr.String())
}
