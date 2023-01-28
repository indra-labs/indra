package indra

import (
	"github.com/indra-labs/lnd/lnd/lnwire"

	"github.com/indra-labs/indra/pkg/crypto/ciph"
	"github.com/indra-labs/indra/pkg/onion"
	"github.com/indra-labs/indra/pkg/onion/layers/crypt"
	"github.com/indra-labs/indra/pkg/onion/layers/reverse"
	"github.com/indra-labs/indra/pkg/types"
	"github.com/indra-labs/indra/pkg/util/slice"
)

func (en *Engine) reverse(on *reverse.Layer, b slice.Bytes,
	c *slice.Cursor, prev types.Onion) {

	var e error
	var on2 types.Onion
	if on.AddrPort.String() == en.Node.AddrPort.String() {
		if on2, e = onion.Peel(b, c); check(e) {
			return
		}
		switch on1 := on2.(type) {
		case *crypt.Layer:
			start := *c - crypt.ReverseLayerLen
			first := *c
			second := first + crypt.ReverseLayerLen
			last := second + crypt.ReverseLayerLen
			log.T.Ln("searching for reverse crypt keys")
			hdr, pld, _, _ := en.FindCloaked(on1.Cloak)
			if hdr == nil || pld == nil {
				log.E.F("failed to find key for %s",
					en.Node.AddrPort.String())
				return
			}
			// We need to find the PayloadPub to match.
			hdrPrv := hdr
			hdrPub := on1.FromPub
			blk := ciph.GetBlock(hdrPrv, hdrPub)
			// Decrypt using the Payload key and header nonce.
			ciph.Encipher(blk, on1.Nonce,
				b[*c:c.Inc(2*crypt.ReverseLayerLen)])
			blk = ciph.GetBlock(pld, hdrPub)
			ciph.Encipher(blk, on1.Nonce, b[*c:])
			// shift the header segment upwards and pad the
			// remainder.
			copy(b[start:first], b[first:second])
			copy(b[first:second], b[second:last])
			copy(b[second:last], slice.NoisePad(crypt.ReverseLayerLen))
			if b[start:start+2].String() != reverse.MagicString {
				// It's for us!
				log.D.Ln("handling response")
				en.handleMessage(BudgeUp(b, last), on)
				break
			}
			sess := en.FindSessionByHeader(hdr)
			if sess != nil {
				log.D.Ln(on.AddrPort.String(), "reverse receive")
				en.DecSession(sess.ID,
					en.RelayRate*lnwire.MilliSatoshi(len(b))/1024/1024)
			}
			en.handleMessage(BudgeUp(b, start), on)
		default:
			// If a reverse is not followed by an onion crypt the
			// message is incorrectly formed, just drop it.
			return
		}
	} else {
		// we need to forward this message onion.
		log.D.Ln("forwarding reverse")
		en.Send(on.AddrPort, b)
	}

}
