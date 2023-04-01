package engine

import (
	"fmt"
	"net/netip"
	"sync"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/ring"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

type Service struct {
	Port      uint16
	RelayRate int
	Transport
}

type Services []*Service

// Node is a representation of a messaging counterparty.
type Node struct {
	ID nonce.ID
	sync.Mutex
	AddrPort  *netip.AddrPort
	Identity  *Keys
	RelayRate int                 // Base relay price mSAT/Mb.
	Services  Services            // Services offered by this peer.
	Load      *ring.BufferLoad    // Relay load.
	Latency   *ring.BufferLatency // Latency to peer.
	Failure   *ring.BufferFailure // Times of tx failure.
	PaymentChan
	Transport
}

const (
	// DefaultSampleBufferSize defines the number of samples for the Load, Latency
	// and Failure ring buffers.
	DefaultSampleBufferSize = 64
	PaymentChanBuffers      = 8
)

// NewNode creates a new Node. A netip.AddrPort is optional if the counterparty is
// not in direct connection. Also, the IdentityPrv node private key can be nil,
// as only the node embedded in a client and not the peer node list has one
// available. The Node for a client's self should use true in the local
// parameter to not initialise the peer state ring buffers as it won't use them.
func NewNode(addr *netip.AddrPort, idPrv *prv.Key,
	tpt Transport, relayRate int, local bool) (n *Node, id nonce.ID) {
	
	id = nonce.NewID()
	n = &Node{
		ID:          id,
		AddrPort:    addr,
		Identity:    MakeKeys(idPrv),
		RelayRate:   relayRate,
		PaymentChan: make(PaymentChan, PaymentChanBuffers),
		Transport:   tpt,
	}
	if !local {
		// These ring buffers are needed to evaluate these metrics for remote
		// peers only.
		n.Load = ring.NewBufferLoad(DefaultSampleBufferSize)
		n.Latency = ring.NewBufferLatency(DefaultSampleBufferSize)
		n.Failure = ring.NewBufferFailure(DefaultSampleBufferSize)
	}
	return
}

func (n *Node) AddService(s *Service) (e error) {
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

func (n *Node) FindService(port uint16) (service *Service) {
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
			n.Services[i].Send(b)
			e = nil
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
