package client

import (
	"net/netip"

	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/crypto/sha256"
	"github.com/indra-labs/indra/pkg/onion"
	"github.com/indra-labs/indra/pkg/onion/layers/crypt"
	"github.com/indra-labs/indra/pkg/onion/layers/directbalance"
	"github.com/indra-labs/indra/pkg/onion/layers/exit"
	"github.com/indra-labs/indra/pkg/onion/layers/forward"
	"github.com/indra-labs/indra/pkg/onion/layers/getbalance"
	"github.com/indra-labs/indra/pkg/onion/layers/reverse"
	"github.com/indra-labs/indra/pkg/util/slice"
	"github.com/indra-labs/lnd/lnd/lnwire"
)

func (cl *Client) SendOnion(ap *netip.AddrPort, o onion.Skins,
	responseHook func(b slice.Bytes)) {
	b := onion.Encode(o.Assemble())
	var billable, accounted []nonce.ID
	var ret nonce.ID
	var last sha256.Hash
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
			s := cl.FindSessionByHeaderPub(on.ToHeaderPub)
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
					cl.AddrPort.String(), "send forward")
				cl.DecSession(s.ID,
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
					cl.DecSession(s.ID,
						s.Services[i].RelayRate*lnwire.MilliSatoshi(len(b)/2)/1024/1024)
					accounted = append(accounted, s.ID)
					break
				}
				billable = append(billable, s.ID)
				last = sha256.Single(on2.Bytes)
				skip = true
			case *getbalance.Layer:
				log.D.Ln("sender: getbalance layer")
				cl.DecSession(s.ID,
					s.RelayRate*lnwire.MilliSatoshi(len(b)/2)/1024/1024)
				last = sha256.Single(s.ID[:])
				billable = append(billable, s.ID)
				skip = true
			}
		case *directbalance.Layer:
			// the immediate previous layer session needs to be accounted.
			switch on3 := o[i-1].(type) {
			case *crypt.Layer:
				s := cl.FindSessionByHeaderPub(on3.ToHeaderPub)
				if s == nil {
					return
				}
				log.D.Ln("sender: directbalance layer")
				cl.DecSession(s.ID,
					s.RelayRate*lnwire.MilliSatoshi(len(b))/1024/1024)
			}
		}
	}
	if responseHook == nil {
		responseHook = func(_ slice.Bytes) {}
	}
	cl.PendingResponses.Add(last, billable, accounted, ret, port, responseHook)
	log.T.Ln("sending out onion")
	cl.Send(ap, b)

}
