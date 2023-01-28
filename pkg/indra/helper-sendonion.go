package indra

import (
	"net/netip"

	"github.com/indra-labs/lnd/lnd/lnwire"

	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/onion"
	"github.com/indra-labs/indra/pkg/onion/layers/crypt"
	"github.com/indra-labs/indra/pkg/onion/layers/directbalance"
	"github.com/indra-labs/indra/pkg/onion/layers/exit"
	"github.com/indra-labs/indra/pkg/onion/layers/forward"
	"github.com/indra-labs/indra/pkg/onion/layers/getbalance"
	"github.com/indra-labs/indra/pkg/onion/layers/reverse"
	"github.com/indra-labs/indra/pkg/util/slice"
)

func (en *Engine) SendOnion(ap *netip.AddrPort, o onion.Skins,
	responseHook func(id nonce.ID, b slice.Bytes)) {
	b := onion.Encode(o.Assemble())
	var billable, accounted []nonce.ID
	var ret nonce.ID
	var last nonce.ID
	var port uint16
	// do client accounting
	skip := false
	for i := range o {
		if skip {
			skip = false
			continue
		}
		switch on := o[i].(type) {
		case *crypt.Layer:
			s := en.FindSessionByHeaderPub(on.ToHeaderPub)
			// The last hop needs no accounting as it's us!
			if i == len(o)-1 {
				// The session used for the last hop is stored, however.
				ret = s.ID
				break
			}
			if s == nil {
				continue
			}
			switch on2 := o[i+1].(type) {
			case *forward.Layer:
				log.D.Ln("sender:",
					en.AddrPort.String(), "send forward")
				en.DecSession(s.ID,
					s.RelayRate*lnwire.MilliSatoshi(len(b))/1024/1024)
				accounted = append(accounted, s.ID)
			case *reverse.Layer:
				billable = append(billable, s.ID)
			case *exit.Layer:
				for i := range s.Services {
					if s.Services[i].Port != on2.Port {
						continue
					}
					port = on2.Port
					log.D.Ln("sender:",
						s.AddrPort.String(), "exit receive")
					en.DecSession(s.ID,
						s.Services[i].RelayRate*lnwire.MilliSatoshi(len(b)/2)/1024/1024)
					accounted = append(accounted, s.ID)
					break
				}
				billable = append(billable, s.ID)
				last = on2.ID
				skip = true
			case *getbalance.Layer:
				log.D.Ln("sender: getbalance layer")
				en.DecSession(s.ID,
					s.RelayRate*lnwire.MilliSatoshi(len(b)/2)/1024/1024)
				last = s.ID
				billable = append(billable, s.ID)
				skip = true
			}
		case *directbalance.Layer:
			// the immediate previous layer session needs to be accounted.
			switch on3 := o[i-1].(type) {
			case *crypt.Layer:
				s := en.FindSessionByHeaderPub(on3.ToHeaderPub)
				if s == nil {
					return
				}
				log.D.Ln("sender: directbalance layer")
				en.DecSession(s.ID,
					s.RelayRate*lnwire.MilliSatoshi(len(b))/1024/1024)
			}
		}
	}
	if responseHook == nil {
		responseHook = func(_ nonce.ID, _ slice.Bytes) {}
	}
	en.Pending.Add(last, billable, accounted, ret, port, responseHook)
	log.T.Ln("sending out onion")
	en.Send(ap, b)

}
