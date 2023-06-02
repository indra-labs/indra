package sess

import (
	"errors"
	"net/netip"

	"github.com/gookit/color"

	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/engine/node"
	"github.com/indra-labs/indra/pkg/engine/responses"
	"github.com/indra-labs/indra/pkg/util/slice"
	"github.com/indra-labs/indra/pkg/util/splice"
)

// Send a message to a peer via their AddrPort.
func (sm *Manager) Send(addr *netip.AddrPort, s *splice.Splice) {
	// first search if we already have the node available with connection open.
	as := addr.String()
	sm.ForEachNode(func(n *node.Node) bool {
		if as == n.AddrPort.String() {
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
func (sm *Manager) SendWithOneHook(ap *netip.AddrPort,
	res *Data, responseHook responses.Callback, p *responses.Pending) {

	if responseHook == nil {
		responseHook = func(_ nonce.ID, _ interface{}, _ slice.Bytes) (e error) {
			log.D.Ln("nil response hook")
			return errors.New("nil response hook")
		}
	}
	log.I.S("res", res.ID)
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
		"to", color.Yellow.Sprint(ap.String()))
	sm.Send(ap, splice.Load(res.B, slice.NewCursor()))
}
