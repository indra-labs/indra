package client

import (
	"net/netip"
	"reflect"
	"sync"
	"time"

	"github.com/Indra-Labs/indra"
	"github.com/Indra-Labs/indra/pkg/ifc"
	"github.com/Indra-Labs/indra/pkg/key/address"
	"github.com/Indra-Labs/indra/pkg/key/prv"
	"github.com/Indra-Labs/indra/pkg/key/pub"
	"github.com/Indra-Labs/indra/pkg/key/signer"
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

const DefaultDeadline = 10 * time.Minute

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
	*signer.KeySet
	qu.C
}

func New(tpt ifc.Transport, hdrPrv *prv.Key, no *node.Node,
	nodes node.Nodes) (c *Client, e error) {

	no.Transport = tpt
	no.HeaderPriv = hdrPrv
	no.HeaderPub = pub.Derive(hdrPrv)
	c = &Client{
		Node:         no,
		Nodes:        nodes,
		ReceiveCache: address.NewReceiveCache(),
		C:            qu.T(),
	}
	c.ReceiveCache.Add(address.NewReceiver(hdrPrv))
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
		cur := slice.NewCursor()
		if onion, e = wire.PeelOnion(b, cur); check(e) {
			break
		}
		switch on := onion.(type) {
		case *cipher.OnionSkin:
			log.I.Ln(reflect.TypeOf(on))
			c.cipher(on, b, cur)
		case *confirmation.OnionSkin:
			log.I.Ln(reflect.TypeOf(on))
			c.confirmation(on, b, cur)
		case *delay.OnionSkin:
			log.I.Ln(reflect.TypeOf(on))
			c.delay(on, b, cur)
		case *exit.OnionSkin:
			log.I.Ln(reflect.TypeOf(on))
			c.exit(on, b, cur)
		case *forward.OnionSkin:
			log.I.Ln(reflect.TypeOf(on))
			c.forward(on, b, cur)
		case *layer.OnionSkin:
			log.I.Ln(reflect.TypeOf(on))
			c.layer(on, b, cur)
		case *noop.OnionSkin:
			log.I.Ln(reflect.TypeOf(on))
			c.noop(on, b, cur)
		case *purchase.OnionSkin:
			log.I.Ln(reflect.TypeOf(on))
			c.purchase(on, b, cur)
		case *reply.OnionSkin:
			log.I.Ln(reflect.TypeOf(on))
			c.reply(on, b, cur)
		case *response.OnionSkin:
			log.I.Ln(reflect.TypeOf(on))
			c.response(on, b, cur)
		case *session.OnionSkin:
			log.I.Ln(reflect.TypeOf(on))
			c.session(on, b, cur)
		case *token.OnionSkin:
			log.I.Ln(reflect.TypeOf(on))
			c.token(on, b, cur)
		default:
			log.I.S("unrecognised packet", b)
		}
	}
	return false
}

func (c *Client) cipher(on *cipher.OnionSkin, b slice.Bytes, cur *slice.Cursor) {
	// This either is in a forward only SendKeys message or we are the buyer
	// and these are our session keys.
}

func (c *Client) confirmation(on *confirmation.OnionSkin, b slice.Bytes,
	cur *slice.Cursor) {
	// This will be an 8 byte nonce that confirms a message passed, ping and
	// cipher onions return these, as they are pure forward messages that
	// send a message one way and the confirmation is the acknowledgement.
	log.I.S(on)
}

func (c *Client) delay(on *delay.OnionSkin, b slice.Bytes, cur *slice.Cursor) {
	// this is a message to hold the message in the buffer until a duration
	// elapses. The accounting for the remainder of the message adds a
	// factor to the effective byte consumption in accordance with the time
	// to be stored.
}

func (c *Client) exit(on *exit.OnionSkin, b slice.Bytes, cur *slice.Cursor) {
	// payload is forwarded to a local port and the response is forwarded
	// back with a reply header.
}

func (c *Client) forward(on *forward.OnionSkin, b slice.Bytes, cur *slice.Cursor) {
	// forward the whole buffer received onwards. Usually there will be a
	// layer.OnionSkin under this which will be unwrapped by the receiver.
	if on.AddrPort.String() == c.Node.AddrPort.String() {
		// it is for us, we want to unwrap the next
		// part.
		log.I.Ln("processing new message")
		c.Node.Send(b[*cur:])
	} else {
		// we need to forward this message onion.
		log.I.Ln("forwarding")
		c.Send(on.AddrPort, b)
	}
}

func (c *Client) layer(on *layer.OnionSkin, b slice.Bytes, cur *slice.Cursor) {
	// this is probably an encrypted layer for us.
	log.I.Ln("decrypting onion skin")
	// log.I.S(on, b[*cur:].ToBytes())
	rcv := c.ReceiveCache.FindCloaked(on.Cloak)
	if rcv == nil {
		log.I.Ln("no matching key found from cloaked key")
		return
	}
	on.Decrypt(rcv.Key, b, cur)
	log.I.S(b[*cur:].ToBytes())
	c.Node.Send(b[*cur:])
}

func (c *Client) noop(on *noop.OnionSkin, b slice.Bytes, cur *slice.Cursor) {
	// this won't happen normally
}

func (c *Client) purchase(on *purchase.OnionSkin, b slice.Bytes, cur *slice.Cursor) {
	// Create a new Session.
	s := &Session{}
	c.Mutex.Lock()
	s.Deadline = time.Now().Add(DefaultDeadline)
	c.Sessions = append(c.Sessions, s)
	c.Mutex.Unlock()
}

func (c *Client) reply(on *reply.OnionSkin, b slice.Bytes, cur *slice.Cursor) {
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

func (c *Client) response(on *response.OnionSkin, b slice.Bytes, cur *slice.Cursor) {
	// Response is a payload from an exit message.
}

func (c *Client) session(s *session.OnionSkin, b slice.Bytes, cur *slice.Cursor) {
	// Session is returned from a Purchase message in the reply layers.
	//
	// Session has a nonce.ID that is given in the last layer of a LN sphinx
	// Bolt 4 onion routed payment path that will cause the seller to
	// activate the accounting onion the two keys it sent back with the
	// nonce, so long as it has not yet expired.
}

func (c *Client) token(t *token.OnionSkin, b slice.Bytes, cur *slice.Cursor) {
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
