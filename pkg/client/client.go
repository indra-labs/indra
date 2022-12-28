package client

import (
	"net/netip"

	"github.com/Indra-Labs/indra"
	"github.com/Indra-Labs/indra/pkg/ifc"
	"github.com/Indra-Labs/indra/pkg/key/address"
	"github.com/Indra-Labs/indra/pkg/key/prv"
	"github.com/Indra-Labs/indra/pkg/key/pub"
	"github.com/Indra-Labs/indra/pkg/node"
	"github.com/Indra-Labs/indra/pkg/slice"
	"github.com/Indra-Labs/indra/pkg/types"
	"github.com/Indra-Labs/indra/pkg/wire"
	"github.com/Indra-Labs/indra/pkg/wire/cipher"
	"github.com/Indra-Labs/indra/pkg/wire/confirmation"
	"github.com/Indra-Labs/indra/pkg/wire/delay"
	"github.com/Indra-Labs/indra/pkg/wire/exit"
	"github.com/Indra-Labs/indra/pkg/wire/forward"
	"github.com/Indra-Labs/indra/pkg/wire/layer"
	"github.com/Indra-Labs/indra/pkg/wire/noop"
	"github.com/Indra-Labs/indra/pkg/wire/purchase"
	"github.com/Indra-Labs/indra/pkg/wire/reply"
	"github.com/Indra-Labs/indra/pkg/wire/response"
	"github.com/Indra-Labs/indra/pkg/wire/session"
	"github.com/Indra-Labs/indra/pkg/wire/token"
	log2 "github.com/cybriq/proc/pkg/log"
	"github.com/cybriq/qu"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

type Client struct {
	*netip.AddrPort
	*node.Node
	node.Nodes
	*address.SendCache
	*address.ReceiveCache
	Circuits
	Sessions
	qu.C
}

func New(tpt ifc.Transport, hdrPrv *prv.Key, no *node.Node,
	nodes node.Nodes) (c *Client, e error) {

	hdrPub := pub.Derive(hdrPrv)
	var n *node.Node
	n, _ = node.New(no.AddrPort, hdrPub, nil, hdrPrv, nil, tpt)
	c = &Client{
		Node:  n,
		Nodes: nodes,
		C:     qu.T(),
	}
	return
}

func (c *Client) Start() {
out:
	for {
		select {
		case <-c.C.Wait():
			c.Cleanup()
			break out
		case b := <-c.Node.Receive():
			// process received message
			var onion types.Onion
			var e error
			cursor := slice.NewCursor()
			if onion, e = wire.PeelOnion(b, cursor); check(e) {
				break
			}
			switch on := onion.(type) {
			case *cipher.OnionSkin:
				// This either is in a forward only SendKeys
				// message or we are the buyer and these are our
				// session keys.
				break

			case *confirmation.OnionSkin:
				// This will be an 8 byte nonce that confirms a
				// message passed, ping and cipher onions return
				// these, as they are pure forward messages that
				// send a message one way and the confirmation
				// is the acknowledgement.
				break

			case *delay.OnionSkin:

				break

			case *exit.OnionSkin:
				// payload is forwarded to a local port and the
				// response is forwarded back with a reply
				// header.
				break

			case *forward.OnionSkin:
				// forward the whole buffer received onwards.
				// Usually there will be an OnionSkin under this
				// which will be unwrapped by the receiver.
				if on.AddrPort.String() == c.Node.AddrPort.String() {
					// it is for us, we want to unwrap the
					// next part.
				} else {
					// we need to forward this message onion.
				}
				break

			case *layer.OnionSkin:
				// this is probably an encrypted layer for us.
				break

			case *noop.OnionSkin:
				// this won't happen normally
				break

			case *purchase.OnionSkin:
				// Purchase requires a return of arbitrary data.
				break

			case *reply.OnionSkin:
				// Reply means another OnionSkin is coming and
				// the payload encryption uses the Payload key.
				break

			case *response.OnionSkin:
				// Response is a payload from an exit message.
				break

			case *session.OnionSkin:
				// Session is returned from a Purchase message
				// in the reply layers.
				//
				// Session has a nonce.ID that is given in the
				// last layer of a LN sphinx Bolt 4 onion routed
				// payment path that will cause the seller to
				// activate the accounting onion the two keys it
				// sent back with the nonce.
				break

			case *token.OnionSkin:
				// not really sure if we are using these.
				break

			default:
				log.I.S("unrecognised packet", b)
			}
		}
	}
}

func (c *Client) Cleanup() {
	// Do cleanup stuff before shutdown.
}

func (c *Client) Shutdown() {
	c.C.Q()
}
