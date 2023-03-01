package relay

import (
	"time"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/ciph"
	"git-indra.lan/indra-labs/indra/pkg/messages/confirm"
	"git-indra.lan/indra-labs/indra/pkg/messages/crypt"
	"git-indra.lan/indra-labs/indra/pkg/messages/delay"
	"git-indra.lan/indra-labs/indra/pkg/messages/forward"
	"git-indra.lan/indra-labs/indra/pkg/messages/magicbytes"
	"git-indra.lan/indra-labs/indra/pkg/messages/response"
	"git-indra.lan/indra-labs/indra/pkg/messages/reverse"
	"git-indra.lan/indra-labs/indra/pkg/messages/session"
	"git-indra.lan/indra-labs/indra/pkg/types"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

func (eng *Engine) confirm(on *confirm.Layer,
	b slice.Bytes, c *slice.Cursor, prev types.Onion) {
	
	// When a confirmation arrives check if it is registered for and run the
	// hook that was registered with it.
	eng.PendingResponses.Delete(on.ID, nil)
}

func (eng *Engine) crypt(on *crypt.Layer, b slice.Bytes,
	c *slice.Cursor, prev types.Onion) {
	
	// this is probably an encrypted crypt for us.
	hdr, _, _, identity := eng.FindCloaked(on.Cloak)
	if hdr == nil {
		log.T.Ln("no matching key found from cloaked key")
		return
	}
	on.ToPriv = hdr
	on.Decrypt(hdr, b, c)
	if identity {
		if string(b[*c:][:magicbytes.Len]) != session.MagicString {
			log.T.Ln("dropping message due to identity key with" +
				" no following session")
			return
		}
		eng.handleMessage(BudgeUp(b, *c), on)
		return
	}
	eng.handleMessage(BudgeUp(b, *c), on)
}

func (eng *Engine) delay(on *delay.Layer, b slice.Bytes,
	c *slice.Cursor, prev types.Onion) {
	
	// this is a message to hold the message in the buffer until a duration
	// elapses. The accounting for the remainder of the message adds a
	// factor to the effective byte consumption in accordance with the time
	// to be stored.
	// todo: accounting
	select {
	case <-time.After(on.Duration):
	}
	eng.handleMessage(BudgeUp(b, *c), on)
}

func (eng *Engine) forward(on *forward.Layer, b slice.Bytes,
	c *slice.Cursor, prev types.Onion) {
	
	// forward the whole buffer received onwards. Usually there will be a
	// crypt.Layer under this which will be unwrapped by the receiver.
	if on.AddrPort.String() == eng.GetLocalNodeAddress().String() {
		// it is for us, we want to unwrap the next part.
		eng.handleMessage(BudgeUp(b, *c), on)
	} else {
		switch on1 := prev.(type) {
		case *crypt.Layer:
			sess := eng.FindSessionByHeader(on1.ToPriv)
			if sess != nil {
				eng.DecSession(sess.ID,
					eng.GetLocalNodeRelayRate()*len(b),
					false, "forward")
			}
		}
		// we need to forward this message onion.
		eng.Send(on.AddrPort, BudgeUp(b, *c))
	}
}

// response is a payload from an exit message.
func (eng *Engine) response(on *response.Layer, b slice.Bytes,
	cur *slice.Cursor, prev types.Onion) {
	
	pending := eng.PendingResponses.Find(on.ID)
	log.T.F("searching for pending ID %x", on.ID)
	if pending != nil {
		for i := range pending.Billable {
			s := eng.FindSession(pending.Billable[i])
			if s != nil {
				typ := "response"
				relayRate := s.RelayRate
				dataSize := len(b)
				switch i {
				case 0, 1:
					dataSize = pending.SentSize
				case 2:
					for j := range s.Services {
						if s.Services[j].Port == on.Port {
							relayRate = s.Services[j].RelayRate / 2
							typ = "exit"
						}
					}
				}
				eng.DecSession(s.ID, relayRate*dataSize, true, typ)
			}
		}
		eng.PendingResponses.Delete(on.ID, on.Bytes)
	}
}

func (eng *Engine) reverse(on *reverse.Layer, b slice.Bytes,
	c *slice.Cursor, prev types.Onion) {
	
	var e error
	var on2 types.Onion
	if on.AddrPort.String() == eng.GetLocalNodeAddress().String() {
		if on2, e = Peel(b, c); check(e) {
			return
		}
		switch on1 := on2.(type) {
		case *crypt.Layer:
			start := *c - crypt.ReverseLayerLen
			first := *c
			second := first + crypt.ReverseLayerLen
			last := second + crypt.ReverseLayerLen
			log.T.Ln("searching for reverse crypt keys")
			hdr, pld, _, _ := eng.FindCloaked(on1.Cloak)
			if hdr == nil || pld == nil {
				log.E.F("failed to find key for %s",
					eng.GetLocalNodeAddress().String())
				return
			}
			// We need to find the PayloadPub to match.
			on1.ToPriv = hdr
			blk := ciph.GetBlock(on1.ToPriv, on1.FromPub)
			// Decrypt using the Payload key and header nonce.
			ciph.Encipher(blk, on1.Nonce,
				b[*c:c.Inc(2*crypt.ReverseLayerLen)])
			blk = ciph.GetBlock(pld, on1.FromPub)
			ciph.Encipher(blk, on1.Nonce, b[*c:])
			// shift the header segment upwards and pad the
			// remainder.
			copy(b[start:first], b[first:second])
			copy(b[first:second], b[second:last])
			copy(b[second:last], slice.NoisePad(crypt.ReverseLayerLen))
			if b[start:start+2].String() != reverse.MagicString {
				// It's for us!
				log.T.Ln("handling response")
				eng.handleMessage(BudgeUp(b, last), on1)
				break
			}
			sess := eng.FindSessionByHeader(hdr)
			if sess != nil {
				eng.DecSession(sess.ID,
					eng.GetLocalNodeRelayRate()*len(b), false, "reverse")
				eng.handleMessage(BudgeUp(b, start), on1)
			}
		default:
			// If a reverse is not followed by an onion crypt the
			// message is incorrectly formed, just drop it.
			return
		}
	} else if prev != nil {
		// we need to forward this message onion.
		log.T.Ln("forwarding reverse")
		eng.Send(on.AddrPort, b)
	} else {
		log.E.Ln("we do not forward nonsense! scoff! snort!")
	}
}
