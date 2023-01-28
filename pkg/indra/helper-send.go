package indra

import (
	"net/netip"

	"github.com/davecgh/go-spew/spew"

	"github.com/indra-labs/indra/pkg/util/slice"
)

// Send a message to a peer via their AddrPort.
func (en *Engine) Send(addr *netip.AddrPort, b slice.Bytes) {
	// first search if we already have the node available with connection
	// open.
	as := addr.String()
	for i := range en.Nodes {
		if as == en.Nodes[i].AddrPort.String() {
			log.T.C(func() string {
				return en.AddrPort.String() +
					" sending to " +
					addr.String() +
					"\n" +
					spew.Sdump(b.ToBytes())
			})
			en.Nodes[i].Transport.Send(b)
			return
		}
	}
	// If we got to here none of the addresses matched, and we need to
	// establish a new peer connection to them, if we know of them (this
	// would usually be the reason this happens).
	log.T.Ln("could not find peer with address", addr.String())
}
