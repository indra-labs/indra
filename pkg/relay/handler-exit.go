package relay

import (
	"time"

	"git-indra.lan/indra-labs/lnd/lnd/lnwire"

	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/onion"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/crypt"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/exit"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/response"
	"git-indra.lan/indra-labs/indra/pkg/types"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

func (eng *Engine) exit(ex *exit.Layer, b slice.Bytes,
	c *slice.Cursor, prev types.Onion) {

	// payload is forwarded to a local port and the result is forwarded
	// back with a reverse header.
	var e error
	var result slice.Bytes
	h := sha256.Single(ex.Bytes)
	log.T.S(h)
	log.T.F("received exit id %x", ex.ID)
	if e = eng.GetLocalNode().SendTo(ex.Port, ex.Bytes); check(e) {
		return
	}
	timer := time.NewTicker(time.Second * 5) // todo: timeout/retries etc
	select {
	case result = <-eng.GetLocalNode().ReceiveFrom(ex.Port):
	case <-timer.C:
	}
	// We need to wrap the result in a message crypt.
	eng.Lock()
	res := onion.Encode(&response.Layer{
		ID:    ex.ID,
		Port:  ex.Port,
		Load:  eng.Load,
		Bytes: result,
	})
	eng.Unlock()
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
				lnwire.MilliSatoshi(len(b)) / 2 / 1024 / 1024
			out := sess.Services[i].RelayRate *
				lnwire.MilliSatoshi(len(rb)) / 2 / 1024 / 1024
			eng.DecSession(sess.ID, in+out, false, "exit")
			break
		}
	}
	eng.handleMessage(rb, ex)
}
