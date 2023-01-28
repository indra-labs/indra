package indra

import (
	"github.com/indra-labs/lnd/lnd/lnwire"

	"github.com/indra-labs/indra/pkg/crypto/key/pub"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/onion"
	"github.com/indra-labs/indra/pkg/onion/layers/crypt"
	"github.com/indra-labs/indra/pkg/onion/layers/directbalance"
	"github.com/indra-labs/indra/pkg/onion/layers/forward"
	"github.com/indra-labs/indra/pkg/onion/layers/magicbytes"
	"github.com/indra-labs/indra/pkg/onion/layers/session"
	"github.com/indra-labs/indra/pkg/types"
	"github.com/indra-labs/indra/pkg/util/slice"
)

func (cl *Engine) crypt(on *crypt.Layer, b slice.Bytes,
	c *slice.Cursor, prev types.Onion) {

	// this is probably an encrypted crypt for us.
	hdr, _, sess, identity := cl.FindCloaked(on.Cloak)
	if hdr == nil {
		log.T.Ln("no matching key found from cloaked key")
		return
	}
	on.ToPriv = hdr
	on.Decrypt(hdr, b, c)
	if identity {
		log.T.F("identity")
		if string(b[*c:][:magicbytes.Len]) != session.MagicString {
			log.T.Ln("dropping message due to identity key with" +
				" no following session")
			return
		}

		cl.handleMessage(BudgeUp(b, *c), on)
		return
	}
	if string(b[*c:][:magicbytes.Len]) == directbalance.MagicString {
		log.D.Ln("directbalance")
		var on1, on2 types.Onion
		var e error
		if on1, e = onion.Peel(b, c); check(e) {
			return
		}
		var balID, confID nonce.ID
		switch db := on1.(type) {
		case *directbalance.Layer:
			log.T.S(cl.AddrPort.String(), db, b[*c:].ToBytes())
			balID = db.ID
			confID = db.ConfID
		default:
			log.T.Ln("malformed/truncated onion")
			return
		}
		if on2, e = onion.Peel(b, c); check(e) {
			return
		}
		switch fwd := on2.(type) {
		case *forward.Layer:
			log.T.S(cl.AddrPort.String(), fwd)
			o := (&onion.Skins{}).
				Forward(fwd.AddrPort).
				Crypt(pub.Derive(hdr), nil, cl.KeySet.Next(), nonce.New(), 0).
				Balance(balID, confID, sess.Remaining)
			rb := onion.Encode(o.Assemble())
			cl.Send(fwd.AddrPort, rb)
			// cl.SendOnion(fwd.AddrPort, o)
			log.D.Ln(cl.AddrPort.String(), "directbalance reply")
			cl.DecSession(sess.ID,
				cl.RelayRate*lnwire.MilliSatoshi(len(b)/2+len(rb)/2)/1024/1024)
			return
		default:
			log.T.Ln("dropping directbalance without following " +
				"forward")
			return
		}
		return
	}
	cl.handleMessage(BudgeUp(b, *c), on)
}
