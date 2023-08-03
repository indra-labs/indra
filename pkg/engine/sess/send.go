package sess

import (
	"errors"
	"github.com/gookit/color"
	"github.com/indra-labs/indra/pkg/util/cryptorand"
	"github.com/multiformats/go-multiaddr"

	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/engine/node"
	"github.com/indra-labs/indra/pkg/engine/responses"
	"github.com/indra-labs/indra/pkg/util/slice"
	"github.com/indra-labs/indra/pkg/util/splice"
)

// Send a message to a peer via their Addresses.
func (sm *Manager) Send(addr multiaddr.Multiaddr, s *splice.Splice) {
	// first search if we already have the node available with connection open.
	as := addr.String()
	sm.ForEachNode(func(n *node.Node) bool {
		// Shouldn't happen, now won't make a problem.
		if len(n.Addresses) < 1 {
			log.T.Ln(n.Identity.Bytes.String() + " node has no addresses!")
			return false
		}
		var addy string
		shuf := make([]string, len(n.Addresses))
		for i := range n.Addresses {
			shuf[i] = n.Addresses[i].String()
		}
		cryptorand.Shuffle(len(shuf), func(i, j int) {
			shuf[i], shuf[j] = shuf[j], shuf[i]
		})
		addy = shuf[0]
		if as == addy {
			log.D.F("%s sending message to %v",
				sm.GetLocalNodeAddressString(), color.Yellow.Sprint(addr))
			n.Transport.Send(s.GetAll())
			return true
		}
		return false
	})
}

// SendWithOneHook is used for onions with only one confirmation hook. Usually
// as returned from PostAcctOnion this is the last, confirmation or response
// layer in an onion.Skins.
func (sm *Manager) SendWithOneHook(ap []multiaddr.Multiaddr,
	res *Data, responseHook responses.Callback, p *responses.Pending) {

	if responseHook == nil {
		responseHook = func(_ nonce.ID, _ interface{}, _ slice.Bytes) (e error) {
			log.D.Ln("nil response hook")
			return errors.New("nil response hook")
		}
	}
	var addr multiaddr.Multiaddr
	addr = sm.nodes[0].PickAddress(sm.Protocols)
	p.Add(responses.ResponseParams{
		ID:       res.ID,
		SentSize: res.B.Len(),
		S:        res.Sessions,
		Billable: res.Billable,
		Ret:      res.Ret,
		Port:     res.Port,
		Callback: responseHook,
		PostAcct: res.PostAcct},
	)
	log.T.Ln(sm.GetLocalNodeAddressString(), "sending out onion", res.ID,
		"to", color.Yellow.Sprint(addr.String()))
	sm.Send(addr, splice.Load(res.B, slice.NewCursor()))
}
