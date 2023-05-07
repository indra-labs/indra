package node

import (
	"fmt"
	"net/netip"
	"sync"
	
	"git-indra.lan/indra-labs/indra"
	"git-indra.lan/indra-labs/indra/pkg/crypto"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/engine/payments"
	"git-indra.lan/indra-labs/indra/pkg/engine/services"
	"git-indra.lan/indra-labs/indra/pkg/engine/tpt"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	fails = log.E.Chk
)

// Node is a representation of a messaging counterparty.
type Node struct {
	ID nonce.ID
	sync.Mutex
	AddrPort  *netip.AddrPort
	Identity  *crypto.Keys
	RelayRate int               // Base relay price mSAT/Mb.
	Services  services.Services // Services offered by this peer.
	payments.Chan
	Transport tpt.Transport
}

const (
	// DefaultSampleBufferSize defines the number of samples for the Load, Latency
	// and Failure ring buffers.
	DefaultSampleBufferSize = 64
	PaymentChanBuffers      = 8
)

// NewNode creates a new Node. The Node for a client's self should use true in
// the local parameter to not initialise the peer state ring buffers as it won't
// use them.
func NewNode(addr *netip.AddrPort, keys *crypto.Keys, tpt tpt.Transport,
	relayRate int) (n *Node, id nonce.ID) {
	
	id = nonce.NewID()
	n = &Node{
		ID:        id,
		AddrPort:  addr,
		Identity:  keys,
		RelayRate: relayRate,
		Chan:      make(payments.Chan, PaymentChanBuffers),
		Transport: tpt,
	}
	return
}

func (n *Node) AddService(s *services.Service) (e error) {
	n.Lock()
	defer n.Unlock()
	for i := range n.Services {
		if n.Services[i].Port == s.Port {
			return fmt.Errorf("service already exists for port %d", s.Port)
		}
	}
	n.Services = append(n.Services, s)
	return
}

func (n *Node) DeleteService(port uint16) {
	n.Lock()
	defer n.Unlock()
	for i := range n.Services {
		if n.Services[i].Port == port {
			if i < 1 {
				n.Services = n.Services[1:]
			} else {
				n.Services = append(n.Services[:i],
					n.Services[i+1:]...)
			}
			return
		}
	}
}

func (n *Node) FindService(port uint16) (svc *services.Service) {
	n.Lock()
	defer n.Unlock()
	for i := range n.Services {
		if n.Services[i].Port == port {
			return n.Services[i]
		}
	}
	return
}

// SendTo delivers a message to a service identified by its port.
func (n *Node) SendTo(port uint16, b slice.Bytes) (e error) {
	n.Lock()
	defer n.Unlock()
	e = fmt.Errorf("%s port not registered %d", n.AddrPort.String(), port)
	for i := range n.Services {
		if n.Services[i].Port == port {
			e = n.Services[i].Send(b)
			return
		}
	}
	return
}

// ReceiveFrom returns the channel that receives messages for a given port.
func (n *Node) ReceiveFrom(port uint16) (b <-chan slice.Bytes) {
	n.Lock()
	defer n.Unlock()
	for i := range n.Services {
		if n.Services[i].Port == port {
			log.T.Ln("receive from")
			b = n.Services[i].Receive()
			return
		}
	}
	return
}
