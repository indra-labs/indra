package relay

import (
	"time"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/signer"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/messages/crypt"
	"git-indra.lan/indra-labs/indra/pkg/messages/exit"
	"git-indra.lan/indra-labs/indra/pkg/messages/response"
	"git-indra.lan/indra-labs/indra/pkg/types"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

// SendExit constructs a message containing an arbitrary payload to a node (3rd
// hop) with a set of 3 ciphers derived from the hidden PayloadPub of the return
// hops that are layered progressively after the Exit message.
//
// The Exit node forwards the packet it receives to the local port specified in
// the Exit message, and then uses the ciphers to encrypt the reply with the
// three ciphers provided, which don't enable it to decrypt the header, only to
// encrypt the payload.
//
// The response is encrypted with the given layers, the ciphers are already
// given in reverse order, so they are decoded in given order to create the
// correct payload encryption to match the PayloadPub combined with the header's
// given public From key.
//
// The header remains a constant size and each node in the Reverse trims off
// their section at the top, moves the next crypt header to the top and pads the
// remainder with noise, so it always looks like the first hop.
func SendExit(port uint16, payload slice.Bytes, id nonce.ID,
	client *Session, s Circuit, ks *signer.KeySet) Skins {
	
	var prvs [3]*prv.Key
	for i := range prvs {
		prvs[i] = ks.Next()
	}
	n := GenNonces(6)
	var returnNonces [3]nonce.IV
	copy(returnNonces[:], n[3:])
	var pubs [3]*pub.Key
	pubs[0] = s[3].PayloadPub
	pubs[1] = s[4].PayloadPub
	pubs[2] = client.PayloadPub
	return Skins{}.
		ReverseCrypt(s[0], ks.Next(), n[0], 3).
		ReverseCrypt(s[1], ks.Next(), n[1], 2).
		ReverseCrypt(s[2], ks.Next(), n[2], 1).
		Exit(port, prvs, pubs, returnNonces, id, payload).
		ReverseCrypt(s[3], prvs[0], n[3], 3).
		ReverseCrypt(s[4], prvs[1], n[4], 2).
		ReverseCrypt(client, prvs[2], n[5], 1)
}

func (eng *Engine) exit(ex *exit.Layer, b slice.Bytes,
	c *slice.Cursor, prev types.Onion) {
	
	// payload is forwarded to a local port and the result is forwarded
	// back with a reverse header.
	var e error
	var result slice.Bytes
	h := sha256.Single(ex.Bytes)
	log.T.S(h)
	log.T.F("%s received exit id %x", eng.GetLocalNodeAddress(), ex.ID)
	if e = eng.SendFromLocalNode(ex.Port, ex.Bytes); check(e) {
		return
	}
	timer := time.NewTicker(time.Second)
	select {
	case result = <-eng.ReceiveToLocalNode(ex.Port):
	case <-timer.C:
	}
	// We need to wrap the result in a message crypt.
	res := Encode(&response.Layer{
		ID:    ex.ID,
		Port:  ex.Port,
		Load:  byte(eng.Load.Load()),
		Bytes: result,
	})
	rb := FormatReply(b[*c:c.Inc(crypt.ReverseHeaderLen)],
		res, ex.Ciphers, ex.Nonces)
	switch on := prev.(type) {
	case *crypt.Layer:
		sess := eng.FindSessionByHeader(on.ToPriv)
		if sess == nil {
			break
		}
		for i := range sess.Services {
			if ex.Port != sess.Services[i].Port {
				continue
			}
			in := sess.Services[i].RelayRate *
				len(b) / 2
			out := sess.Services[i].RelayRate *
				len(rb) / 2
			eng.DecSession(sess.ID, in+out, false, "exit")
			break
		}
	}
	eng.handleMessage(rb, ex)
}

func (eng *Engine) SendExit(port uint16, msg slice.Bytes, id nonce.ID,
	target *Session, hook Callback) {
	
	hops := []byte{0, 1, 2, 3, 4, 5}
	s := make(Sessions, len(hops))
	s[2] = target
	se := eng.SelectHops(hops, s)
	var c Circuit
	copy(c[:], se)
	o := SendExit(port, msg, id, se[len(se)-1], c, eng.KeySet)
	log.D.Ln("sending out exit onion")
	res := eng.PostAcctOnion(o)
	eng.SendWithOneHook(c[0].AddrPort, res, hook)
}

func (eng *Engine) MakeExit(port uint16, msg slice.Bytes, id nonce.ID,
	exit *Session) (c Circuit, o Skins) {
	
	hops := []byte{0, 1, 2, 3, 4, 5}
	s := make(Sessions, len(hops))
	s[2] = exit
	se := eng.SelectHops(hops, s)
	copy(c[:], se)
	o = SendExit(port, msg, id, se[len(se)-1], c, eng.KeySet)
	return
}

func (eng *Engine) SendExitNew(c Circuit, o Skins, hook Callback) {
	log.D.Ln("sending out exit onion")
	res := eng.PostAcctOnion(o)
	eng.SendWithOneHook(c[0].AddrPort, res, hook)
}
