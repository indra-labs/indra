package relay

import (
	"net/netip"
	"runtime"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/relay/messages/balance"
	"git-indra.lan/indra-labs/indra/pkg/relay/messages/confirm"
	"git-indra.lan/indra-labs/indra/pkg/relay/messages/crypt"
	"git-indra.lan/indra-labs/indra/pkg/relay/messages/exit"
	"git-indra.lan/indra-labs/indra/pkg/relay/messages/forward"
	"git-indra.lan/indra-labs/indra/pkg/relay/messages/getbalance"
	"git-indra.lan/indra-labs/indra/pkg/relay/messages/hiddenservice"
	"git-indra.lan/indra-labs/indra/pkg/relay/messages/reverse"
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

type SendData struct {
	b         slice.Bytes
	sessions  Sessions
	billable  []nonce.ID
	ret, last nonce.ID
	port      uint16
	postAcct  []func()
}

// PostAcctOnion takes a slice of Skins and calculates their costs and
// the list of sessions inside them and attaches accounting operations to
// apply when the associated confirmation(s) or response hooks are executed.
func (eng *Engine) PostAcctOnion(o Skins) (res SendData) {
	res.b = Encode(o.Assemble())
	// do client accounting
	skip := false
	for i := range o {
		if skip {
			skip = false
			continue
		}
		switch on := o[i].(type) {
		case *crypt.Layer:
			s := eng.FindSessionByHeaderPub(on.ToHeaderPub)
			if s == nil {
				continue
			}
			res.sessions = append(res.sessions, s)
			// The last hop needs no accounting as it's us!
			if i == len(o)-1 {
				// The session used for the last hop is stored, however.
				res.ret = s.ID
				res.billable = append(res.billable, s.ID)
				break
			}
			switch on2 := o[i+1].(type) {
			case *forward.Layer:
				res.billable = append(res.billable, s.ID)
				res.postAcct = append(res.postAcct,
					func() { eng.DecSession(s.ID, s.RelayRate*len(res.b), true, "forward") })
			case *hiddenservice.Layer:
				res.last = on2.ID
				res.billable = append(res.billable, s.ID)
				skip = true
			case *reverse.Layer:
				res.billable = append(res.billable, s.ID)
			case *exit.Layer:
				for j := range s.Services {
					if s.Services[j].Port != on2.Port {
						continue
					}
					res.port = on2.Port
					res.postAcct = append(res.postAcct,
						func() { eng.DecSession(s.ID, s.Services[j].RelayRate*len(res.b)/2, true, "exit") })
					break
				}
				res.billable = append(res.billable, s.ID)
				res.last = on2.ID
				skip = true
			case *getbalance.Layer:
				res.last = s.ID
				res.billable = append(res.billable, s.ID)
				skip = true
			}
		case *confirm.Layer:
			res.last = on.ID
		case *balance.Layer:
			res.last = on.ID
		}
	}
	return
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
	eng.PendingResponses.Add(PendingResponseParams{
		res.last,
		len(res.b),
		res.sessions,
		res.billable,
		res.ret,
		res.port,
		responseHook,
		res.postAcct})
	log.T.Ln("sending out onion")
	eng.Send(ap, res.b)
}
