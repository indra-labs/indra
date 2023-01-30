package relay

import (
	"git-indra.lan/indra-labs/lnd/lnd/lnwire"

	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/onion"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/balance"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/crypt"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/directbalance"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/forward"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/magicbytes"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/session"
	"git-indra.lan/indra-labs/indra/pkg/types"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

func (en *Engine) crypt(on *crypt.Layer, b slice.Bytes,
	c *slice.Cursor, prev types.Onion) {

	// this is probably an encrypted crypt for us.
	hdr, _, sess, identity := en.FindCloaked(on.Cloak)
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

		en.handleMessage(BudgeUp(b, *c), on)
		return
	}
	if string(b[*c:][:magicbytes.Len]) == directbalance.MagicString {
		var on1, on2 types.Onion
		var e error
		if on1, e = onion.Peel(b, c); check(e) {
			return
		}
		var balID, confID nonce.ID
		switch db := on1.(type) {
		case *directbalance.Layer:
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
			o := (&onion.Skins{}).
				Forward(fwd.AddrPort).
				Crypt(pub.Derive(hdr), nil, en.KeySet.Next(), nonce.New(), 0).
				Balance(balID, confID, sess.Remaining)
			oo := o.Assemble()
			// This is a little more complicated as we need to decrement the
			// amount before sending out the balance.
			en.DecSession(sess.ID,
				(en.RelayRate*lnwire.MilliSatoshi(len(b)+oo.Len())/2)/1024/1024,
				false, "directbalance")
			o[2].(*balance.Layer).MilliSatoshi = sess.Remaining
			rb := onion.Encode(oo)
			en.Send(fwd.AddrPort, rb)
			// en.SendOnion(fwd.AddrPort, o)
			return
		default:
			log.T.Ln("dropping directbalance without following " +
				"forward")
			return
		}
		return
	}
	en.handleMessage(BudgeUp(b, *c), on)
}
