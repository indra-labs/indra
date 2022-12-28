package client

import (
	"net/netip"
	"sync"

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
	sync.Mutex
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

// Start a single thread of the Client.
func (c *Client) Start() {
out:
	for {
		if c.runner() {
			break out
		}
	}
}

func (c *Client) runner() (out bool) {
	select {
	case <-c.C.Wait():
		c.Cleanup()
		out = true
		break
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
			c.cipher(on, b)
			break

		case *confirmation.OnionSkin:
			c.confirmation(on, b)
			break

		case *delay.OnionSkin:
			c.delay(on, b)
			break

		case *exit.OnionSkin:
			c.exit(on, b)
			break

		case *forward.OnionSkin:
			c.forward(on, b)
			break

		case *layer.OnionSkin:
			c.layer(on, b)
			break

		case *noop.OnionSkin:
			c.noop(on, b)
			break

		case *purchase.OnionSkin:
			c.purchase(on, b)
			break

		case *reply.OnionSkin:
			c.reply(on, b)
			break

		case *response.OnionSkin:
			c.response(on, b)
			break

		case *session.OnionSkin:
			c.session(on, b)
			break

		case *token.OnionSkin:
			c.token(on, b)
			break

		default:
			log.I.S("unrecognised packet", b)
		}
	}
	return false
}

func (c *Client) cipher(on *cipher.OnionSkin, b slice.Bytes) {
	// This either is in a forward only SendKeys message or we are the buyer
	// and these are our session keys.
}

func (c *Client) confirmation(on *confirmation.OnionSkin, b slice.Bytes) {
	// This will be an 8 byte nonce that confirms a message passed, ping and
	// cipher onions return these, as they are pure forward messages that
	// send a message one way and the confirmation is the acknowledgement.
}

func (c *Client) delay(on *delay.OnionSkin, b slice.Bytes) {
	// this is a message to hold the message in the buffer until a time
	// elapses. The accounting for the remainder of the message adds a
	// factor to the effective byte consumption in accordance with the time
	// to be stored.
}

func (c *Client) exit(on *exit.OnionSkin, b slice.Bytes) {
	// payload is forwarded to a local port and the response is forwarded
	// back with a reply header.
}

func (c *Client) forward(on *forward.OnionSkin, b slice.Bytes) {
	// forward the whole buffer received onwards. Usually there will be an
	// OnionSkin under this which will be unwrapped by the receiver.
	if on.AddrPort.String() == c.Node.AddrPort.String() {
		// it is for us, we want to unwrap the next
		// part.
		c.Node.Send(b)
	} else {
		// we need to forward this message onion.
		c.Send(on.AddrPort, b)
	}
}

func (c *Client) layer(on *layer.OnionSkin, b slice.Bytes) {
	// this is probably an encrypted layer for us.
}

func (c *Client) noop(on *noop.OnionSkin, b slice.Bytes) {
	// this won't happen normally
}

func (c *Client) purchase(on *purchase.OnionSkin, b slice.Bytes) {
	// Create a new Session.
	s := &Session{}
	c.Mutex.Lock()
	c.Sessions = append(c.Sessions, s)
	c.Mutex.Unlock()
}

func (c *Client) reply(on *reply.OnionSkin, b slice.Bytes) {
	// Reply means another OnionSkin is coming and the payload encryption
	// uses the Payload key.
	if on.AddrPort.String() == c.Node.AddrPort.String() {
		// it is for us, we want to unwrap the next part.
		c.Node.Send(b)
	} else {
		// we need to forward this message onion.
		c.Send(on.AddrPort, b)
	}

}

func (c *Client) response(on *response.OnionSkin, b slice.Bytes) {
	// Response is a payload from an exit message.
}

func (c *Client) session(s *session.OnionSkin, b slice.Bytes) {
	// Session is returned from a Purchase message in the reply layers.
	//
	// Session has a nonce.ID that is given in the last layer of a LN sphinx
	// Bolt 4 onion routed payment path that will cause the seller to
	// activate the accounting onion the two keys it sent back with the
	// nonce, so long as it has not yet expired.
}

func (c *Client) token(t *token.OnionSkin, b slice.Bytes) {
	// not really sure if we are using these.
	return
}

func (c *Client) Cleanup() {
	// Do cleanup stuff before shutdown.
}

func (c *Client) Shutdown() {
	c.C.Q()
}

func (c *Client) Send(addr *netip.AddrPort, b slice.Bytes) {
	// first search if we already have the node available with connection
	// open.
	as := addr.String()
	for i := range c.Nodes {
		if as == c.Nodes[i].Addr {
			c.Nodes[i].Transport.Send(b)
			return
		}
	}
	// If we got to here none of the addresses matched, and we need to
	// establish a new peer connection to them.

}
