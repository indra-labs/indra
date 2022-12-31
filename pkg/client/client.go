package client

import (
	"net/netip"
	"reflect"
	"sync"
	"time"

	"github.com/Indra-Labs/indra"
	"github.com/Indra-Labs/indra/pkg/ciph"
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
	"go.uber.org/atomic"
)

const (
	DefaultDeadline = 10 * time.Minute
	ReplyHeaderLen  = 3*reverse.Len + 3*layer.Len
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
	*confirm.Confirms
	sync.Mutex
	*signer.KeySet
	atomic.Bool
	qu.C
}

func New(tpt ifc.Transport, hdrPrv *prv.Key, no *node.Node,
	nodes node.Nodes) (c *Client, e error) {

	no.Transport = tpt
	no.HeaderPrv = hdrPrv
	no.HeaderPub = pub.Derive(hdrPrv)
	c = &Client{
		Confirms:     confirm.NewConfirms(),
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
		// log.I.Ln("received message")
		var onion types.Onion
		var e error
		c := slice.NewCursor()
		if onion, e = wire.PeelOnion(b, c); check(e) {
			break
		}
		switch on := onion.(type) {
		case *cipher.OnionSkin:
			log.I.Ln(reflect.TypeOf(on))
			cl.cipher(on, b, c)
		case *confirm.OnionSkin:
			log.I.Ln(reflect.TypeOf(on))
			cl.confirm(on, b, c)
		case *delay.OnionSkin:
			log.I.Ln(reflect.TypeOf(on))
			cl.delay(on, b, c)
		case *exit.OnionSkin:
			log.I.Ln(reflect.TypeOf(on))
			cl.exit(on, b, c)
		case *forward.OnionSkin:
			log.I.Ln(reflect.TypeOf(on))
			cl.forward(on, b, c)
		case *layer.OnionSkin:
			log.I.Ln(reflect.TypeOf(on))
			cl.layer(on, b, c)
		case *noop.OnionSkin:
			log.I.Ln(reflect.TypeOf(on))
			cl.noop(on, b, c)
		case *purchase.OnionSkin:
			log.I.Ln(reflect.TypeOf(on))
			cl.purchase(on, b, c)
		case *reverse.OnionSkin:
			log.I.Ln(reflect.TypeOf(on))
			cl.reverse(on, b, c)
		case *response.OnionSkin:
			log.I.Ln(reflect.TypeOf(on))
			cl.response(on, b, c)
		case *session.OnionSkin:
			log.I.Ln(reflect.TypeOf(on))
			cl.session(on, b, c)
		case *token.OnionSkin:
			log.I.Ln(reflect.TypeOf(on))
			cl.token(on, b, c)
		default:
			log.I.S("unrecognised packet", b)
		}
	}
	return
}

func (cl *Client) cipher(on *cipher.OnionSkin, b slice.Bytes, c *slice.Cursor) {
	// This either is in a forward only SendKeys message or we are the buyer
	// and these are our session keys.
	log.I.S(on)
	b = append(b[*c:], slice.NoisePad(int(*c))...)
	cl.Node.Send(b)
}

func (cl *Client) confirm(on *confirm.OnionSkin, b slice.Bytes,
	c *slice.Cursor) {
	// When a confirm arrives check if it is registered for and run
	// the hook that was registered with it.
	cl.Confirms.Confirm(on.ID)
}

func (cl *Client) RegisterConfirmation(hook confirm.Hook, on *confirm.OnionSkin) {

	cl.Confirms.Add(&confirm.Callback{
		ID:    on.ID,
		Time:  time.Now(),
		Hook:  hook,
		Onion: on,
	})
}

func (cl *Client) delay(on *delay.OnionSkin, b slice.Bytes, cur *slice.Cursor) {
	// this is a message to hold the message in the buffer until a duration
	// elapses. The accounting for the remainder of the message adds a
	// factor to the effective byte consumption in accordance with the time
	// to be stored.
}

func (cl *Client) exit(on *exit.OnionSkin, b slice.Bytes, cur *slice.Cursor) {
	// payload is forwarded to a local port and the response is forwarded
	// back with a reverse header.
}

func (cl *Client) forward(on *forward.OnionSkin, b slice.Bytes, c *slice.Cursor) {
	// forward the whole buffer received onwards. Usually there will be a
	// layer.OnionSkin under this which will be unwrapped by the receiver.
	if on.AddrPort.String() == cl.Node.AddrPort.String() {
		// it is for us, we want to unwrap the next
		// part.
		log.I.Ln("processing new message", *c, cl.AddrPort.String())
		b = append(b[*c:], slice.NoisePad(int(*c))...)
		cl.Node.Send(b)
	} else {
		// we need to forward this message onion.
		log.I.Ln("forwarding")
		cl.Send(on.AddrPort, b)
	}
}

func (cl *Client) layer(on *layer.OnionSkin, b slice.Bytes, c *slice.Cursor) {
	// this is probably an encrypted layer for us.
	log.I.Ln("decrypting onion skin")
	// log.I.S(on, b[*c:].ToBytes())
	rcv := cl.ReceiveCache.FindCloaked(on.Cloak)
	if rcv == nil {
		log.I.Ln("no matching key found from cloaked key")
		return
	}
	on.Decrypt(rcv.Key, b, c)
	b = append(b[*c:], slice.NoisePad(int(*c))...)
	// log.I.S(b.ToBytes())
	cl.Node.Send(b)
}

func (cl *Client) noop(on *noop.OnionSkin, b slice.Bytes, c *slice.Cursor) {
	// this won't happen normally
}

func (cl *Client) purchase(on *purchase.OnionSkin, b slice.Bytes, c *slice.Cursor) {
	// Create a new Session.
	s := NewSession(nonce.NewID(), on.NBytes, DefaultDeadline, cl.KeySet)
	se := &session.OnionSkin{
		ID:         s.ID,
		HeaderKey:  s.HeaderKey.Key,
		PayloadKey: s.PayloadKey.Key,
		Onion:      &noop.OnionSkin{},
	}
	cl.Mutex.Lock()
	cl.Sessions.Add(s)
	cl.Mutex.Unlock()
	header := b[*c:c.Inc(ReplyHeaderLen)]
	rb := make(slice.Bytes, ReplyHeaderLen+session.Len)
	cur := slice.NewCursor()
	copy(rb[*cur:cur.Inc(ReplyHeaderLen)], header[:ReplyHeaderLen])
	start := *cur
	se.Encode(rb, cur)
	for i := range on.Ciphers {
		blk := ciph.BlockFromHash(on.Ciphers[i])
		ciph.Encipher(blk, on.Nonces[i], rb[start:])
	}
	cl.Node.Send(rb)
}

func (cl *Client) reverse(on *reverse.OnionSkin, b slice.Bytes,
	c *slice.Cursor) {

	log.I.Ln("reverse")
	// Reply means another OnionSkin is coming and the payload encryption
	// uses the Payload key.
	if on.AddrPort.String() == cl.Node.AddrPort.String() {
		log.I.Ln("it's for us")
		// it is for us, we want to unwrap the next part.
		cl.Node.Send(b)
	} else {
		// we need to forward this message onion.
		// cl.Send(on.AddrPort, b)
		log.I.Ln("we haven't processed it yet")
	}

}

func (cl *Client) response(on *response.OnionSkin, b slice.Bytes, cur *slice.Cursor) {
	// Response is a payload from an exit message.
}

func (cl *Client) session(s *session.OnionSkin, b slice.Bytes, cur *slice.Cursor) {
	// Session is returned from a Purchase message in the reverse layers.
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

func (cl *Client) Cleanup() {
	// Do cleanup stuff before shutdown.
}

func (cl *Client) Shutdown() {
	if cl.Bool.Load() {
		return
	}
	log.I.Ln("shutting down client", cl.Node.AddrPort.String())
	cl.Bool.Store(true)
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
