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
	"github.com/Indra-Labs/indra/pkg/nonce"
	"github.com/Indra-Labs/indra/pkg/slice"
	"github.com/Indra-Labs/indra/pkg/types"
	"github.com/Indra-Labs/indra/pkg/wire"
	"github.com/Indra-Labs/indra/pkg/wire/cipher"
	"github.com/Indra-Labs/indra/pkg/wire/confirm"
	"github.com/Indra-Labs/indra/pkg/wire/delay"
	"github.com/Indra-Labs/indra/pkg/wire/exit"
	"github.com/Indra-Labs/indra/pkg/wire/forward"
	"github.com/Indra-Labs/indra/pkg/wire/layer"
	"github.com/Indra-Labs/indra/pkg/wire/noop"
	"github.com/Indra-Labs/indra/pkg/wire/purchase"
	"github.com/Indra-Labs/indra/pkg/wire/response"
	"github.com/Indra-Labs/indra/pkg/wire/reverse"
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
	Confirms
	sync.Mutex
	*signer.KeySet
	qu.C
}
type ConfirmHook func(cf *confirm.OnionSkin)

type ConfirmCallback struct {
	nonce.ID
	Hook ConfirmHook
}

type Confirms []ConfirmCallback

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
func (cl *Client) Start() {
out:
	for {
		if cl.runner() {
			break out
		}
	}
}

func (cl *Client) runner() (out bool) {
	select {
	case <-cl.C.Wait():
		cl.Cleanup()
		out = true
		break
	case b := <-cl.Node.Receive():
		// process received message
		var onion types.Onion
		var e error
		c := slice.NewCursor()
		if onion, e = wire.PeelOnion(b, c); check(e) {
			break
		}
		switch on := onion.(type) {
		case *cipher.OnionSkin:
			log.T.Ln(reflect.TypeOf(on))
			cl.cipher(on, b, c)
		case *confirm.OnionSkin:
			log.T.Ln(reflect.TypeOf(on))
			cl.confirm(on, b, c)
		case *delay.OnionSkin:
			log.T.Ln(reflect.TypeOf(on))
			cl.delay(on, b, c)
		case *exit.OnionSkin:
			log.T.Ln(reflect.TypeOf(on))
			cl.exit(on, b, c)
		case *forward.OnionSkin:
			log.T.Ln(reflect.TypeOf(on))
			cl.forward(on, b, c)
		case *layer.OnionSkin:
			log.T.Ln(reflect.TypeOf(on))
			cl.layer(on, b, c)
		case *noop.OnionSkin:
			log.T.Ln(reflect.TypeOf(on))
			cl.noop(on, b, c)
		case *purchase.OnionSkin:
			log.T.Ln(reflect.TypeOf(on))
			cl.purchase(on, b, c)
		case *reverse.OnionSkin:
			log.T.Ln(reflect.TypeOf(on))
			cl.reply(on, b, c)
		case *response.OnionSkin:
			log.T.Ln(reflect.TypeOf(on))
			cl.response(on, b, c)
		case *session.OnionSkin:
			log.T.Ln(reflect.TypeOf(on))
			cl.session(on, b, c)
		case *token.OnionSkin:
			log.T.Ln(reflect.TypeOf(on))
			cl.token(on, b, c)
		default:
			log.I.S("unrecognised packet", b)
		}
	}
	return false
}

func (cl *Client) cipher(on *cipher.OnionSkin, b slice.Bytes, cur *slice.Cursor) {
	// This either is in a forward only SendKeys message or we are the buyer
	// and these are our session keys.
}

func (cl *Client) confirm(on *confirm.OnionSkin, b slice.Bytes,
	c *slice.Cursor) {
	// This will be an 8 byte nonce that confirms a message passed, ping and
	// cipher onions return these, as they are pure forward messages that
	// send a message one way and the confirmation is the acknowledgement.
	// log.I.S(on)
	for i := range cl.Confirms {
		// log.I.S(cl.Confirms[i].ID, on.ID)
		if on.ID == cl.Confirms[i].ID {
			cl.Confirms[i].Hook(on)
		}
	}
}

func (cl *Client) delay(on *delay.OnionSkin, b slice.Bytes, cur *slice.Cursor) {
	// this is a message to hold the message in the buffer until a duration
	// elapses. The accounting for the remainder of the message adds a
	// factor to the effective byte consumption in accordance with the time
	// to be stored.
}

func (cl *Client) exit(on *exit.OnionSkin, b slice.Bytes, cur *slice.Cursor) {
	// payload is forwarded to a local port and the response is forwarded
	// back with a reply header.
}

func (cl *Client) forward(on *forward.OnionSkin, b slice.Bytes, c *slice.Cursor) {
	// forward the whole buffer received onwards. Usually there will be a
	// layer.OnionSkin under this which will be unwrapped by the receiver.
	if on.AddrPort.String() == cl.Node.AddrPort.String() {
		// it is for us, we want to unwrap the next
		// part.
		log.T.Ln("processing new message", *c)
		b = append(b[*c:], slice.NoisePad(int(*c))...)
		cl.Node.Send(b)
	} else {
		// we need to forward this message onion.
		log.T.Ln("forwarding")
		cl.Send(on.AddrPort, b)
	}
}

func (cl *Client) layer(on *layer.OnionSkin, b slice.Bytes, c *slice.Cursor) {
	// this is probably an encrypted layer for us.
	log.T.Ln("decrypting onion skin")
	// log.I.S(on, b[*c:].ToBytes())
	rcv := cl.ReceiveCache.FindCloaked(on.Cloak)
	if rcv == nil {
		log.T.Ln("no matching key found from cloaked key")
		return
	}
	on.Decrypt(rcv.Key, b, c)
	b = append(b[*c:], slice.NoisePad(int(*c))...)
	log.T.S(b.ToBytes())
	cl.Node.Send(b)
}

func (cl *Client) noop(on *noop.OnionSkin, b slice.Bytes, c *slice.Cursor) {
	// this won't happen normally
}

func (cl *Client) purchase(on *purchase.OnionSkin, b slice.Bytes, c *slice.Cursor) {
	// Create a new Session.
	s := &Session{}
	cl.Mutex.Lock()
	s.Deadline = time.Now().Add(DefaultDeadline)
	cl.Sessions = append(cl.Sessions, s)
	cl.Mutex.Unlock()
}

func (cl *Client) reply(on *reverse.OnionSkin, b slice.Bytes, c *slice.Cursor) {
	// Reply means another OnionSkin is coming and the payload encryption
	// uses the Payload key.
	if on.AddrPort.String() == cl.Node.AddrPort.String() {
		// it is for us, we want to unwrap the next part.
		cl.Node.Send(b)
	} else {
		// we need to forward this message onion.
		cl.Send(on.AddrPort, b)
	}

}

func (cl *Client) response(on *response.OnionSkin, b slice.Bytes, cur *slice.Cursor) {
	// Response is a payload from an exit message.
}

func (cl *Client) session(s *session.OnionSkin, b slice.Bytes, cur *slice.Cursor) {
	// Session is returned from a Purchase message in the reply layers.
	//
	// Session has a nonce.ID that is given in the last layer of a LN sphinx
	// Bolt 4 onion routed payment path that will cause the seller to
	// activate the accounting onion the two keys it sent back with the
	// nonce, so long as it has not yet expired.
}

func (cl *Client) token(t *token.OnionSkin, b slice.Bytes, cur *slice.Cursor) {
	// not really sure if we are using these.
	return
}

func (cl *Client) RegisterConfirmation(id nonce.ID, hook ConfirmHook) {
	cl.Mutex.Lock()
	cl.Confirms = append(cl.Confirms, ConfirmCallback{
		ID:   id,
		Hook: hook,
	})
	cl.Mutex.Unlock()
}

func (cl *Client) Cleanup() {
	// Do cleanup stuff before shutdown.
}

func (cl *Client) Shutdown() {
	cl.C.Q()
}

func (cl *Client) Send(addr *netip.AddrPort, b slice.Bytes) {
	// first search if we already have the node available with connection
	// open.
	as := addr.String()
	for i := range cl.Nodes {
		if as == cl.Nodes[i].Addr {
			cl.Nodes[i].Transport.Send(b)
			return
		}
	}
	// If we got to here none of the addresses matched, and we need to
	// establish a new peer connection to them.

}
