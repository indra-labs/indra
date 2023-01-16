package client

import (
	"fmt"
	"reflect"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/indra-labs/indra/pkg/ciph"
	"github.com/indra-labs/indra/pkg/node"
	"github.com/indra-labs/indra/pkg/nonce"
	"github.com/indra-labs/indra/pkg/sha256"
	"github.com/indra-labs/indra/pkg/slice"
	"github.com/indra-labs/indra/pkg/types"
	"github.com/indra-labs/indra/pkg/wire"
	"github.com/indra-labs/indra/pkg/wire/balance"
	"github.com/indra-labs/indra/pkg/wire/confirm"
	"github.com/indra-labs/indra/pkg/wire/delay"
	"github.com/indra-labs/indra/pkg/wire/exit"
	"github.com/indra-labs/indra/pkg/wire/forward"
	"github.com/indra-labs/indra/pkg/wire/getbalance"
	"github.com/indra-labs/indra/pkg/wire/layer"
	"github.com/indra-labs/indra/pkg/wire/magicbytes"
	"github.com/indra-labs/indra/pkg/wire/noop"
	"github.com/indra-labs/indra/pkg/wire/response"
	"github.com/indra-labs/indra/pkg/wire/reverse"
	"github.com/indra-labs/indra/pkg/wire/session"
	"github.com/indra-labs/indra/pkg/wire/token"
)

func recLog(on types.Onion, b slice.Bytes, cl *Client) func() string {
	return func() string {
		return cl.AddrPort.String() +
			" received " +
			fmt.Sprint(reflect.TypeOf(on)) + "\n" +
			spew.Sdump(b.ToBytes())
	}
}

func (cl *Client) runner() (out bool) {
	log.T.C(func() string {
		return cl.AddrPort.String() +
			" awaiting message"
	})
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
		case *balance.OnionSkin:
			log.T.C(recLog(on, b, cl))
			cl.balance(on, b, c)
		case *confirm.OnionSkin:
			log.T.C(recLog(on, b, cl))
			cl.confirm(on, b, c)
		case *delay.OnionSkin:
			log.T.C(recLog(on, b, cl))
			cl.delay(on, b, c)
		case *exit.OnionSkin:
			log.T.C(recLog(on, b, cl))
			cl.exit(on, b, c)
		case *forward.OnionSkin:
			log.T.C(recLog(on, b, cl))
			cl.forward(on, b, c)
		case *getbalance.OnionSkin:
			log.T.C(recLog(on, b, cl))
			cl.getBalance(on, b, c)
		case *layer.OnionSkin:
			log.T.C(recLog(on, b, cl))
			cl.layer(on, b, c)
		case *noop.OnionSkin:
			log.T.C(recLog(on, b, cl))
			cl.noop(on, b, c)
		case *reverse.OnionSkin:
			log.T.C(recLog(on, b, cl))
			cl.reverse(on, b, c)
		case *response.OnionSkin:
			log.T.C(recLog(on, b, cl))
			cl.response(on, b, c)
		case *session.OnionSkin:
			log.T.C(recLog(on, b, cl))
			cl.session(on, b, c)
		case *token.OnionSkin:
			log.T.C(recLog(on, b, cl))
			cl.token(on, b, c)
		default:
			log.I.S("unrecognised packet", b)
		}
	case p := <-cl.PaymentChan:
		log.T.S("incoming payment", cl.AddrPort.String(), p)
		topUp := false
		cl.IterateSessions(func(s *node.Session) bool {
			if s.Preimage == p.Preimage {
				s.AddBytes(p.Amount)
				topUp = true
				log.T.F("topping up %x with %d mSat",
					s.ID, p.Amount)
				return true
			}
			return false
		})
		if !topUp {
			cl.PendingPayments = cl.PendingPayments.Add(p)
			log.T.F("awaiting session keys for preimage %x",
				p.Preimage)
		}
	}
	return
}

func BudgeUp(b slice.Bytes, start slice.Cursor) (o slice.Bytes) {
	o = b
	copy(o, o[start:])
	copy(o[len(o)-int(start):], slice.NoisePad(int(start)))
	return
}

func (cl *Client) confirm(on *confirm.OnionSkin,
	b slice.Bytes, c *slice.Cursor) {

	// When a confirm arrives check if it is registered for and run
	// the hook that was registered with it.
	log.T.S(cl.Confirms)
	cl.Confirms.Confirm(on.ID)
}

func (cl *Client) balance(on *balance.OnionSkin,
	b slice.Bytes, c *slice.Cursor) {

	cl.IterateSessions(func(s *node.Session) bool {
		if s.ID == on.ID {
			log.T.F("received balance %x for session %x",
				on.MilliSatoshi, on.ID)
			s.Remaining = on.MilliSatoshi
			return true
		}
		return false
	})
}

func (cl *Client) delay(on *delay.OnionSkin, b slice.Bytes,
	c *slice.Cursor) {

	// this is a message to hold the message in the buffer until a duration
	// elapses. The accounting for the remainder of the message adds a
	// factor to the effective byte consumption in accordance with the time
	// to be stored.
	// todo: accounting
	select {
	case <-time.After(on.Duration):
	}
	cl.Node.Send(BudgeUp(b, *c))
}

func (cl *Client) exit(on *exit.OnionSkin, b slice.Bytes,
	c *slice.Cursor) {

	// payload is forwarded to a local port and the result is forwarded
	// back with a reverse header.
	var e error
	var result slice.Bytes
	if e = cl.SendTo(on.Port, on.Bytes); check(e) {
		return
	}
	timer := time.NewTicker(time.Second * 5)
	select {
	case result = <-cl.ReceiveFrom(on.Port):
	case <-timer.C:
	}
	// We need to wrap the result in a message layer.
	res := wire.EncodeOnion(&response.OnionSkin{
		Hash:  sha256.Single(on.Bytes),
		Bytes: result,
	})
	rb := FormatReply(b[*c:c.Inc(ReverseHeaderLen)],
		res, on.Ciphers, on.Nonces)
	cl.Node.Send(rb)
}

func (cl *Client) forward(on *forward.OnionSkin, b slice.Bytes,
	c *slice.Cursor) {

	// forward the whole buffer received onwards. Usually there will be a
	// layer.OnionSkin under this which will be unwrapped by the receiver.
	if on.AddrPort.String() == cl.Node.AddrPort.String() {
		// it is for us, we want to unwrap the next part.
		// cl.Node.Send(append(b[*c:], slice.NoisePad(int(*c))...))
		cl.Node.Send(BudgeUp(b, *c))
	} else {
		// we need to forward this message onion.
		cl.Send(on.AddrPort, b)
	}
}

func (cl *Client) getBalance(on *getbalance.OnionSkin,
	b slice.Bytes, c *slice.Cursor) {

	var found bool
	var bal *balance.OnionSkin
	cl.IterateSessions(func(s *node.Session) bool {
		if s.ID == on.ID {
			bal = &balance.OnionSkin{
				ID:           on.ID,
				MilliSatoshi: s.Remaining,
			}
			found = true
			return true
		}
		return false
	})
	if !found {
		return
	}
	rb := FormatReply(b[*c:c.Inc(ReverseHeaderLen)],
		wire.EncodeOnion(bal), on.Ciphers, on.Nonces)
	cl.Node.Send(rb)
}

func FormatReply(header, res slice.Bytes, ciphers [3]sha256.Hash,
	nonces [3]nonce.IV) (rb slice.Bytes) {

	rb = make(slice.Bytes, ReverseHeaderLen+len(res))
	cur := slice.NewCursor()
	copy(rb[*cur:cur.Inc(ReverseHeaderLen)], header[:ReverseHeaderLen])
	copy(rb[ReverseHeaderLen:], res)
	start := *cur
	for i := range ciphers {
		blk := ciph.BlockFromHash(ciphers[i])
		ciph.Encipher(blk, nonces[2-i], rb[start:])
	}
	return
}

func (cl *Client) layer(on *layer.OnionSkin, b slice.Bytes,
	c *slice.Cursor) {

	// this is probably an encrypted layer for us.
	hdr, _, _, identity := cl.FindCloaked(on.Cloak)
	if hdr == nil {
		log.T.Ln("no matching key found from cloaked key")
		return
	}
	on.Decrypt(hdr, b, c)
	if identity {
		if string(b[*c:][:magicbytes.Len]) != session.MagicString {
			log.T.Ln("dropping message due to identity key with" +
				" no following session")
			return
		}
	}
	cl.Node.Send(BudgeUp(b, *c))
}

func (cl *Client) noop(on *noop.OnionSkin, b slice.Bytes,
	c *slice.Cursor) {

	// this won't happen normally
}

func (cl *Client) reverse(on *reverse.OnionSkin, b slice.Bytes,
	c *slice.Cursor) {

	var e error
	var onion types.Onion
	if on.AddrPort.String() == cl.Node.AddrPort.String() {
		if onion, e = wire.PeelOnion(b, c); check(e) {
			return
		}
		switch on1 := onion.(type) {
		case *layer.OnionSkin:
			start := *c - ReverseLayerLen
			first := *c
			second := first + ReverseLayerLen
			last := second + ReverseLayerLen
			log.T.Ln("searching for reverse layer keys")
			hdr, pld, _, _ := cl.FindCloaked(on1.Cloak)
			if hdr == nil || pld == nil {
				log.E.F("failed to find key for %s",
					cl.Node.AddrPort.String())
				return
			}
			// We need to find the PayloadPub to match.
			hdrPrv := hdr
			hdrPub := on1.FromPub
			blk := ciph.GetBlock(hdrPrv, hdrPub)
			// Decrypt using the Payload key and header nonce.
			ciph.Encipher(blk, on1.Nonce,
				b[*c:c.Inc(2*ReverseLayerLen)])
			blk = ciph.GetBlock(pld, hdrPub)
			ciph.Encipher(blk, on1.Nonce, b[*c:])
			// shift the header segment upwards and pad the
			// remainder.
			copy(b[start:first], b[first:second])
			copy(b[first:second], b[second:last])
			copy(b[second:last], slice.NoisePad(ReverseLayerLen))
			if b[start:start+2].String() != reverse.MagicString {
				cl.Node.Send(b[last:])
				break
			}
			cl.Node.Send(BudgeUp(b, start))
		default:
			// If a reverse is not followed by an onion layer the
			// message is incorrectly formed, just drop it.
			return
		}
	} else {
		// we need to forward this message onion.
		cl.Send(on.AddrPort, b)
	}

}

func (cl *Client) response(on *response.OnionSkin, b slice.Bytes,
	cur *slice.Cursor) {

	// Response is a payload from an exit message.
	cl.Hooks.Find(on.Hash, on.Bytes)
}

func (cl *Client) session(on *session.OnionSkin, b slice.Bytes,
	c *slice.Cursor) {

	log.T.C(func() string {
		return fmt.Sprint("incoming session",
			spew.Sdump(on.PreimageHash()))
	})
	pi := cl.PendingPayments.FindPreimage(on.PreimageHash())
	if pi != nil {
		ss := node.NewSession(pi.ID,
			cl.Node, pi.Amount, on.Header, on.Payload, on.Hop)
		cl.AddSession(ss)
		log.T.F("Adding session %x\n", pi.ID)
		cl.PendingPayments = cl.PendingPayments.Delete(pi.Preimage)
		cl.Node.Send(BudgeUp(b, *c))
	} else {
		log.T.Ln("dropping session message without payment")
	}
}

func (cl *Client) token(t *token.OnionSkin, b slice.Bytes,
	cur *slice.Cursor) {

	// not really sure if we are using these.
	return
}
