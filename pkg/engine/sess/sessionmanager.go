package sess

import (
	"fmt"
	"net/netip"
	"sync"

	"github.com/gookit/color"
	"github.com/lightningnetwork/lnd/lnwire"

	"github.com/indra-labs/indra"
	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/crypto/sha256"
	"github.com/indra-labs/indra/pkg/engine/node"
	"github.com/indra-labs/indra/pkg/engine/payments"
	"github.com/indra-labs/indra/pkg/engine/services"
	"github.com/indra-labs/indra/pkg/engine/sessions"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/util/slice"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	fails = log.E.Chk
)

func (sc CircuitCache) Add(s *sessions.Data) CircuitCache {
	var sce *sessions.Circuit
	var exists bool
	if sce, exists = sc[s.Node.ID]; !exists {
		sce = &sessions.Circuit{}
		sce[s.Hop] = s
		sc[s.Node.ID] = sce
		return sc
	}
	sc[s.Node.ID][s.Hop] = s
	return sc
}

type (
	// A CircuitCache stores each of the 5 hops of a peer node.
	CircuitCache map[nonce.ID]*sessions.Circuit
)
type (
	// Manager is a session manager for Indra, handling sessions and services.
	Manager struct {
		nodes           []*node.Node
		PendingPayments payments.PendingPayments
		sessions.Sessions
		CircuitCache
		sync.Mutex
	}
)

// AddNodes adds a Node to a Nodes.
func (sm *Manager) AddNodes(nn ...*node.Node) {
	sm.Lock()
	defer sm.Unlock()
	sm.nodes = append(sm.nodes, nn...)
}

// PendingPayment accessors. For the same reason as the sessions, pending
// payments need to be accessed only with the node's mutex locked.

// AddPendingPayment adds a received incoming payment message to await the
// session keys.
func (sm *Manager) AddPendingPayment(np *payments.Payment) {
	sm.Lock()
	defer sm.Unlock()
	log.D.F("%s adding pending payment %s for %v",
		sm.nodes[0].AddrPort.String(), np.ID,
		np.Amount)
	sm.PendingPayments = sm.PendingPayments.Add(np)
}

// AddServiceToLocalNode adds a service to the local node.
func (sm *Manager) AddServiceToLocalNode(s *services.Service) (e error) {
	sm.Lock()
	defer sm.Unlock()
	return sm.GetLocalNode().AddService(s)
}

// AddSession adds a session to the session cache.
func (sm *Manager) AddSession(s *sessions.Data) {
	sm.Lock()
	defer sm.Unlock()
	// check for dupes
	for i := range sm.Sessions {
		if sm.Sessions[i].Header.Bytes == s.Header.Bytes {
			log.D.F("refusing to add duplicate session Keys %x", s.Header.Bytes)
			return
		}
	}
	sm.Sessions = append(sm.Sessions, s)
	// Hop 5, the return session( s) are not added to the CircuitCache as they
	// are not Billable and are only related to the node of the Engine.
	if s.Hop < 5 {
		sm.CircuitCache = sm.CircuitCache.Add(s)
	}
}

// ClearPendingPayments is used only for debugging, removing all pending
// payments, making the engine forget about payments it received.
func (sm *Manager) ClearPendingPayments() {
	log.D.Ln("clearing pending payments")
	sm.PendingPayments = sm.PendingPayments[:0]
}

// ClearSessions is used only for debugging, removing all but the first session,
// which is the engine's initial return session.
func (sm *Manager) ClearSessions() {
	log.D.Ln("clearing sessions")
	sm.Sessions = sm.Sessions[:1]
}

// DecSession decrements credit (mSat) on a session.
func (sm *Manager) DecSession(id crypto.PubBytes, msats int, sender bool,
	typ string) bool {

	sess := sm.FindSessionByPubkey(id)
	if sess != nil {
		sm.Lock()
		defer sm.Unlock()
		return sess.DecSats(lnwire.MilliSatoshi(msats/1024/1024),
			sender, typ)
	}
	return false
}

// DeleteNodeAndSessions deletes a node and all the sessions for it.
func (sm *Manager) DeleteNodeAndSessions(id nonce.ID) {
	sm.Lock()
	defer sm.Unlock()
	var exists bool
	// If the node exists its Keys is in the CircuitCache.
	if _, exists = sm.CircuitCache[id]; !exists {
		return
	}
	delete(sm.CircuitCache, id)
	// ProcessAndDelete from the nodes list.
	for i := range sm.nodes {
		if sm.nodes[i].ID == id {
			sm.nodes = append(sm.nodes[:i], sm.nodes[i+1:]...)
			break
		}
	}
	var found []int
	// Locate all the sessions with the node in them.
	for i := range sm.Sessions {
		if sm.Sessions[i].Node.ID == id {
			found = append(found, i)
		}
	}
	// Create a new Sessions slice and add the ones not in the found list.
	temp := make(sessions.Sessions, 0, len(sm.Sessions)-len(found))
	for i := range sm.Sessions {
		for j := range found {
			if i != found[j] {
				temp = append(temp, sm.Sessions[i])
				break
			}
		}
	}
	// Place the new Sessions slice in place of the old.
	sm.Sessions = temp
}

// DeleteNodeByAddrPort deletes a node identified by a netip.AddrPort.
func (sm *Manager) DeleteNodeByAddrPort(ip *netip.AddrPort) (e error) {
	sm.Lock()
	defer sm.Unlock()
	e = fmt.Errorf("node with ip %v not found", ip)
	for i := range sm.nodes {
		if sm.nodes[i].AddrPort.String() == ip.String() {
			sm.nodes = append(sm.nodes[:i], sm.nodes[i+1:]...)
			e = nil
			break
		}
	}
	return
}

// DeleteNodeByID deletes a node identified by an Keys.
func (sm *Manager) DeleteNodeByID(ii nonce.ID) (e error) {
	sm.Lock()
	defer sm.Unlock()
	e = fmt.Errorf("id %x not found", ii)
	for i := range sm.nodes {
		if sm.nodes[i].ID == ii {
			sm.nodes = append(sm.nodes[:i], sm.nodes[i+1:]...)
			return
		}
	}
	return
}

// DeletePendingPayment deletes a pending payment by the preimage hash.
func (sm *Manager) DeletePendingPayment(preimage sha256.Hash) {
	sm.Lock()
	defer sm.Unlock()
	sm.PendingPayments = sm.PendingPayments.Delete(preimage)
}
func (sm *Manager) DeleteSession(id crypto.PubBytes) {
	sm.Lock()
	defer sm.Unlock()
	for i := range sm.Sessions {
		if sm.Sessions[i].Header.Bytes == id {
			// ProcessAndDelete from Data cache.
			sm.CircuitCache[sm.Sessions[i].Node.ID][sm.Sessions[i].Hop] = nil
			// ProcessAndDelete from
			sm.Sessions = append(sm.Sessions[:i], sm.Sessions[i+1:]...)
		}
	}
}

// FindCloaked searches the client identity key and the sessions for a match. It
// returns the session as well, though not all users of this function will need
// this.
func (sm *Manager) FindCloaked(clk crypto.CloakedPubKey) (hdr *crypto.Prv,
	pld *crypto.Prv, sess *sessions.Data, identity bool) {

	var b crypto.Blinder
	copy(b[:], clk[:crypto.BlindLen])
	hash := crypto.Cloak(b, sm.GetLocalNodeIdentityBytes())
	if hash == clk {
		log.T.F("encrypted to identity key")
		hdr = sm.GetLocalNodeIdentityPrv()
		// there is no payload key for the node, only in
		identity = true
		return
	}
	sm.IterateSessions(func(s *sessions.Data) (stop bool) {
		hash = crypto.Cloak(b, s.Header.Bytes)
		if hash == clk {
			hdr = s.Header.Prv
			pld = s.Payload.Prv
			sess = s
			return true
		}
		return
	})
	return
}

// FindNodeByAddrPort searches for a Node by netip.AddrPort.
func (sm *Manager) FindNodeByAddrPort(id *netip.AddrPort) (no *node.Node) {
	sm.Lock()
	defer sm.Unlock()
	for _, nn := range sm.nodes {
		if nn.AddrPort.String() == id.String() {
			no = nn
			break
		}
	}
	return
}

// FindNodeByID searches for a Node by Keys.
func (sm *Manager) FindNodeByID(i nonce.ID) (no *node.Node) {
	sm.Lock()
	defer sm.Unlock()
	for _, nn := range sm.nodes {
		if nn.ID == i {
			no = nn
			break
		}
	}
	return
}

// FindNodeByIdentity searches for a Node by netip.AddrPort.
func (sm *Manager) FindNodeByIdentity(id *crypto.Pub) (no *node.Node) {
	sm.Lock()
	defer sm.Unlock()
	for _, nn := range sm.nodes {
		if nn.Identity.Pub.Equals(id) {
			no = nn
			break
		}
	}
	return
}

// FindNodeByIndex returns the node at a given position in the array.
func (sm *Manager) FindNodeByIndex(i int) (no *node.Node) {
	sm.Lock()
	defer sm.Unlock()
	return sm.nodes[i]
}

// FindPendingPayment searches for a pending payment with the matching ID.
func (sm *Manager) FindPendingPayment(id nonce.ID) (pp *payments.Payment) {
	sm.Lock()
	defer sm.Unlock()
	return sm.PendingPayments.Find(id)
}

// FindPendingPreimage searches for a pending payment with e matching preimage.
func (sm *Manager) FindPendingPreimage(pi sha256.Hash) (pp *payments.Payment) {
	log.T.F("searching preimage %s", pi)
	sm.Lock()
	defer sm.Unlock()
	return sm.PendingPayments.FindPreimage(pi)
}

// FindSessionByHeader searches for a session with a matching header private key.
func (sm *Manager) FindSessionByHeader(prvKey *crypto.Prv) *sessions.Data {
	sm.Lock()
	defer sm.Unlock()
	for i := range sm.Sessions {
		if sm.Sessions[i].Header.Prv.Key.Equals(&prvKey.Key) {
			return sm.Sessions[i]
		}
	}
	return nil
}

// FindSessionByHeaderPub searches for a session with a matching header public
// key.
func (sm *Manager) FindSessionByHeaderPub(pubKey *crypto.Pub) *sessions.Data {
	sm.Lock()
	defer sm.Unlock()
	for i := range sm.Sessions {
		if sm.Sessions[i].Header.Pub.Equals(pubKey) {
			return sm.Sessions[i]
		}
	}
	return nil
}

// FindSessionByPubkey searches for a session with a matching public key.
func (sm *Manager) FindSessionByPubkey(id crypto.PubBytes) *sessions.Data {
	sm.Lock()
	defer sm.Unlock()
	for i := range sm.Sessions {
		if sm.Sessions[i].Header.Bytes == id {
			return sm.Sessions[i]
		}
	}
	return nil
}

// FindSessionPreimage searches for a session with a matching preimage hash.
func (sm *Manager) FindSessionPreimage(pr sha256.Hash) *sessions.Data {
	sm.Lock()
	defer sm.Unlock()
	for i := range sm.Sessions {
		if sm.Sessions[i].Preimage == pr {
			return sm.Sessions[i]
		}
	}
	return nil
}

// ForEachNode runs a function over the slice of nodes with the mutex locked,
// and terminates when the function returns true.
//
// Do not call any Manager methods above inside this function or there
// will be a mutex double locking panic, except GetLocalNode.
func (sm *Manager) ForEachNode(fn func(n *node.Node) bool) {
	sm.Lock()
	defer sm.Unlock()
	for i := range sm.nodes {
		if i == 0 {
			continue
		}
		if fn(sm.nodes[i]) {
			return
		}
	}
}

// GetLocalNode returns the engine's local Node.
func (sm *Manager) GetLocalNode() *node.Node { return sm.nodes[0] }

// GetLocalNodeAddress returns the AddrPort of the local node.
func (sm *Manager) GetLocalNodeAddress() (addr *netip.AddrPort) {
	//sm.Lock()
	//defer sm.Unlock()
	return sm.GetLocalNode().AddrPort
}

// GetLocalNodeAddressString returns the string form of the local node address.
func (sm *Manager) GetLocalNodeAddressString() (s string) {
	return color.Yellow.Sprint(sm.GetLocalNodeAddress())
}

// GetLocalNodeIdentityBytes returns the public key bytes of the local node.
func (sm *Manager) GetLocalNodeIdentityBytes() (ident crypto.PubBytes) {
	sm.Lock()
	defer sm.Unlock()
	return sm.GetLocalNode().Identity.Bytes
}

// GetLocalNodeIdentityPrv returns the identity private key of the local node.
func (sm *Manager) GetLocalNodeIdentityPrv() (ident *crypto.Prv) {
	sm.Lock()
	defer sm.Unlock()
	return sm.GetLocalNode().Identity.Prv
}

// GetLocalNodePaymentChan returns the engine's local Node Chan.
func (sm *Manager) GetLocalNodePaymentChan() payments.Chan {
	return sm.nodes[0].Chan
}

// GetLocalNodeRelayRate returns the relay rate for the local node.
func (sm *Manager) GetLocalNodeRelayRate() (rate int) {
	sm.Lock()
	defer sm.Unlock()
	return sm.GetLocalNode().RelayRate
}

// GetNodeCircuit gets the set of 5 sessions associated with a node with a given
// ID.
func (sm *Manager) GetNodeCircuit(id nonce.ID) (sce *sessions.Circuit,
	exists bool) {

	sm.Lock()
	defer sm.Unlock()
	sce, exists = sm.CircuitCache[id]
	return
}

// GetSessionByIndex returns the session with the given index in the main session
// cache.
func (sm *Manager) GetSessionByIndex(i int) (s *sessions.Data) {
	sm.Lock()
	defer sm.Unlock()
	if len(sm.Sessions) > i {
		s = sm.Sessions[i]
	}
	return
}

// GetSessionsAtHop returns all of the sessions designated for a given hop in the
// circuit.
func (sm *Manager) GetSessionsAtHop(hop byte) (s sessions.Sessions) {
	sm.Lock()
	defer sm.Unlock()
	for i := range sm.Sessions {
		if sm.Sessions[i].Hop == hop {
			s = append(s, sm.Sessions[i])
		}
	}
	return
}

// IncSession adds an amount of mSat to the balance of a session.
func (sm *Manager) IncSession(id crypto.PubBytes, msats lnwire.MilliSatoshi,
	sender bool, typ string) {

	sess := sm.FindSessionByPubkey(id)
	if sess != nil {
		sm.Lock()
		defer sm.Unlock()
		sess.IncSats(msats, sender, typ)
	}
}

// IterateSessionCache calls a function for each entry in the CircuitCache
// that provides also access to the related node.
//
// Do not call Manager methods within this function.
func (sm *Manager) IterateSessionCache(fn func(n *node.Node,
	c *sessions.Circuit) bool) {

	sm.Lock()
	defer sm.Unlock()
out:
	for i := range sm.CircuitCache {
		for j := range sm.nodes {
			if sm.nodes[j].ID == i {
				if fn(sm.nodes[j], sm.CircuitCache[i]) {
					break out
				}
				break
			}
		}
	}
}

// IterateSessions calls a function for each entry in the Sessions slice.
//
// Do not call Manager methods within this function.
func (sm *Manager) IterateSessions(fn func(s *sessions.Data) bool) {
	sm.Lock()
	defer sm.Unlock()
	for i := range sm.Sessions {
		if fn(sm.Sessions[i]) {
			break
		}
	}
}

// NodesLen returns the length of a Nodes.
func (sm *Manager) NodesLen() int {
	sm.Lock()
	defer sm.Unlock()
	return len(sm.nodes)
}

// ReceiveToLocalNode returns a channel that will receive messages for the local
// node, that arrived from the internet.
func (sm *Manager) ReceiveToLocalNode() <-chan slice.Bytes {
	sm.Lock()
	defer sm.Unlock()
	return sm.GetLocalNode().Transport.Receive()
}

// SendFromLocalNode delivers a message to a local service.
func (sm *Manager) SendFromLocalNode(port uint16,
	b slice.Bytes) (e error) {

	sm.Lock()
	defer sm.Unlock()
	return sm.GetLocalNode().SendTo(port, b)
}

// SetLocalNode sets the engine's local Node.
func (sm *Manager) SetLocalNode(n *node.Node) {
	sm.Lock()
	defer sm.Unlock()
	sm.nodes[0] = n
}

// SetLocalNodeAddress changes the local node address.
func (sm *Manager) SetLocalNodeAddress(addr *netip.AddrPort) {
	sm.Lock()
	defer sm.Unlock()
	sm.GetLocalNode().AddrPort = addr
}

// UpdateSessionCache reads the main Sessions cache and populates the
// CircuitCache where circuits are aggregated.
func (sm *Manager) UpdateSessionCache() {
	sm.Lock()
	defer sm.Unlock()
	// First we create CircuitCache entries for all existing nodes.
	for i := range sm.nodes {
		_, exists := sm.CircuitCache[sm.nodes[i].ID]
		if !exists {
			sm.CircuitCache[sm.nodes[i].ID] = &sessions.Circuit{}
		}
	}
	// Place all sessions in their slots respective to their node.
	for _, v := range sm.Sessions {
		sm.CircuitCache[v.Node.ID][v.Hop] = v
	}
}

// NewSessionManager creates a new session manager.
func NewSessionManager() *Manager {
	return &Manager{
		CircuitCache:    make(CircuitCache),
		PendingPayments: make(payments.PendingPayments, 0),
	}
}
