package node

import (
	"fmt"
	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/engine/payments"
	"github.com/indra-labs/indra/pkg/engine/services"
	"github.com/indra-labs/indra/pkg/engine/tpt"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/util/slice"
	"net/netip"
	"sync"
)

const (
	// PaymentChanBuffers is the default number of buffers used in a payment channel.
	PaymentChanBuffers = 8
)

var (
	log   = log2.GetLogger()
	fails = log.E.Chk
)

// Node is a representation of a messaging counterparty.
type Node struct {
	ID nonce.ID
	sync.Mutex
	AddrPort  *netip.AddrPort
	Identity  *crypto.Keys
	RelayRate uint32               // Base relay price mSAT/Mb.
	Services  services.Services // Services offered by this peer.
	payments.PayChan
	Transport tpt.Transport
}

// NewNode creates a new Node. The transport should be from either dialing out or
// a peer dialing in and the self model does not need to do this.
func NewNode(addr *netip.AddrPort, keys *crypto.Keys, tpt tpt.Transport,
	relayRate uint32) (n *Node, id nonce.ID) {
	id = nonce.NewID()
	n = &Node{
		ID:        id,
		AddrPort:  addr,
		Identity:  keys,
		RelayRate: relayRate,
		PayChan:   make(payments.PayChan, PaymentChanBuffers),
		Transport: tpt,
	}
	return
}

// AddService adds a service to a Node.
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

// DeleteService removes a service from a Node.
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

// FindService searches for a local service with a given port number.
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
