// Package node is a specification for the in-memory metadata related to an indra network peer.
//
// This structure aggregates the address, identity keys, relay rate, services and payments channel for the peer.
package node

import (
	"fmt"
	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/engine/payments"
	"github.com/indra-labs/indra/pkg/engine/protocols"
	"github.com/indra-labs/indra/pkg/engine/services"
	"github.com/indra-labs/indra/pkg/engine/tpt"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/util/multikey"
	"github.com/indra-labs/indra/pkg/util/slice"
	"github.com/multiformats/go-multiaddr"
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

	// ID is a unique identifier used internally for references.
	ID nonce.ID

	// Mutex to stop concurrent read/write.
	*sync.Mutex

	// Addresses is the network addresses a node is listening to.
	//
	// These can be multiple, but for reasons of complexity, they are filtered by the available protocols for the session manager, ie ip4 always, ip6 sometimes.
	Addresses []multiaddr.Multiaddr

	// Identity is the crypto.Keys identifying the node on the Indra network.
	Identity *crypto.Keys

	// RelayRate is the base relay price mSAT/Mb.
	RelayRate uint32

	// Services offered by this peer.
	Services services.Services

	// Load is the current level of utilisation of the node's resources.
	Load byte

	// PayChan is the channel that payments to this node are sent/received on (internal/external node).
	payments.PayChan

	// Transport is the way to contact the node. Sending messages on this channel go
	// to the dispatcher to be segmented and delivered, or conversely assembled and
	// received.
	Transport tpt.Transport
}

// NewNode creates a new Node. The transport should be from either dialing out or
// a peer dialing in and the self model does not need to do this.
func NewNode(addr []multiaddr.Multiaddr, keys *crypto.Keys, tpt tpt.Transport,
	relayRate uint32) (n *Node, id nonce.ID) {
	id = nonce.NewID()
	n = &Node{
		ID:        id,
		Mutex:     &sync.Mutex{},
		Addresses: addr,
		Identity:  keys,
		RelayRate: relayRate,
		PayChan:   make(payments.PayChan, PaymentChanBuffers),
		Transport: tpt,
	}
	return
}

func (n *Node) PickAddress(p protocols.NetworkProtocols) (ma multiaddr.Multiaddr) {
	var tmp []multiaddr.Multiaddr
	var e error
	for _, temp := range n.Addresses {
		if p&protocols.IP4 != 0 {
			if _, e = temp.ValueForProtocol(multiaddr.P_IP4); e == nil {
				tmp = append(tmp, multikey.AddKeyToMultiaddr(temp, n.Identity.Pub))
			}
		}
		if p&protocols.IP6 != 0 {
			if _, e = temp.ValueForProtocol(multiaddr.P_IP6); e == nil {
				tmp = append(tmp, multikey.AddKeyToMultiaddr(temp, n.Identity.Pub))
			}
		}
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
	e = fmt.Errorf("%v port not registered %d", n.Addresses, port)
	for i := range n.Services {
		if n.Services[i].Port == port {
			e = n.Services[i].Send(b)
			return
		}
	}
	e = fmt.Errorf("%v port not registered %d", n.Addresses, port)
	return
}
