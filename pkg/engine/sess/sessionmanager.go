package sess

import (
	"fmt"
	"net/netip"
	"sync"
	
	"git-indra.lan/indra-labs/lnd/lnd/lnwire"
	"github.com/gookit/color"
	
	"git-indra.lan/indra-labs/indra"
	"git-indra.lan/indra-labs/indra/pkg/crypto"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/engine/node"
	"git-indra.lan/indra-labs/indra/pkg/engine/payments"
	"git-indra.lan/indra-labs/indra/pkg/engine/services"
	"git-indra.lan/indra-labs/indra/pkg/engine/sessions"
	"git-indra.lan/indra-labs/indra/pkg/engine/transport"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	fails = log.E.Chk
)

type Manager struct {
	nodes           []*node.Node
	Listeners       []*transport.Listener
	PendingPayments payments.PendingPayments
	sessions.Sessions
	SessionCache
	sync.Mutex
}

func NewSessionManager(listeners ...*transport.Listener) *Manager {
	return &Manager{
		SessionCache:    make(SessionCache),
		PendingPayments: make(payments.PendingPayments, 0),
		Listeners:       listeners,
	}
}

// FindCloaked searches the client identity key and the sessions for a match. It
// returns the session as well, though not all users of this function will need
// this.
func (sm *Manager) FindCloaked(clk crypto.PubKey) (hdr *crypto.Prv,
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

func (sm *Manager) IncSession(id crypto.PubBytes, msats lnwire.MilliSatoshi,
	sender bool, typ string) {
	
	sess := sm.FindSession(id)
	if sess != nil {
		sm.Lock()
		defer sm.Unlock()
		sess.IncSats(msats, sender, typ)
	}
}

func (sm *Manager) DecSession(id crypto.PubBytes, msats int, sender bool,
	typ string) bool {
	
	sess := sm.FindSession(id)
	if sess != nil {
		sm.Lock()
		defer sm.Unlock()
		return sess.DecSats(lnwire.MilliSatoshi(msats/1024/1024),
			sender, typ)
	}
	return false
}

func (sm *Manager) GetNodeCircuit(id nonce.ID) (sce *sessions.Circuit,
	exists bool) {
	
	sm.Lock()
	defer sm.Unlock()
	sce, exists = sm.SessionCache[id]
	return
}

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
	// Hop 5, the return session( s) are not added to the SessionCache as they
	// are not Billable and are only related to the node of the Engine.
	if s.Hop < 5 {
		sm.SessionCache = sm.SessionCache.Add(s)
	}
}
func (sm *Manager) FindSession(id crypto.PubBytes) *sessions.Data {
	sm.Lock()
	defer sm.Unlock()
	for i := range sm.Sessions {
		if sm.Sessions[i].Header.Bytes == id {
			return sm.Sessions[i]
		}
	}
	return nil
}
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
func (sm *Manager) DeleteSession(id crypto.PubBytes) {
	sm.Lock()
	defer sm.Unlock()
	for i := range sm.Sessions {
		if sm.Sessions[i].Header.Bytes == id {
			// ProcessAndDelete from Data cache.
			sm.SessionCache[sm.Sessions[i].Node.ID][sm.Sessions[i].Hop] = nil
			// ProcessAndDelete from 
			sm.Sessions = append(sm.Sessions[:i], sm.Sessions[i+1:]...)
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

// IterateSessionCache calls a function for each entry in the SessionCache
// that provides also access to the related node.
//
// Do not call Manager methods within this function.
func (sm *Manager) IterateSessionCache(fn func(n *node.Node,
	c *sessions.Circuit) bool) {
	
	sm.Lock()
	defer sm.Unlock()
out:
	for i := range sm.SessionCache {
		for j := range sm.nodes {
			if sm.nodes[j].ID == i {
				if fn(sm.nodes[j], sm.SessionCache[i]) {
					break out
				}
				break
			}
		}
	}
}

func (sm *Manager) GetSessionByIndex(i int) (s *sessions.Data) {
	sm.Lock()
	defer sm.Unlock()
	if len(sm.Sessions) > i {
		s = sm.Sessions[i]
	}
	return
}

func (sm *Manager) DeleteNodeAndSessions(id nonce.ID) {
	sm.Lock()
	defer sm.Unlock()
	var exists bool
	// If the node exists its Keys is in the SessionCache.
	if _, exists = sm.SessionCache[id]; !exists {
		return
	}
	delete(sm.SessionCache, id)
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

// NodesLen returns the length of a Nodes.
func (sm *Manager) NodesLen() int {
	sm.Lock()
	defer sm.Unlock()
	return len(sm.nodes)
}

// GetLocalNode returns the engine's local Node.
func (sm *Manager) GetLocalNode() *node.Node { return sm.nodes[0] }

// GetLocalNodePaymentChan returns the engine's local Node Chan.
func (sm *Manager) GetLocalNodePaymentChan() payments.Chan {
	return sm.nodes[0].Chan
}

func (sm *Manager) GetLocalNodeAddress() (addr *netip.AddrPort) {
	// sm.Lock()
	// defer sm.Unlock()
	return sm.GetLocalNode().AddrPort
}

func (sm *Manager) GetLocalNodeAddressString() (s string) {
	return color.Yellow.Sprint(sm.GetLocalNodeAddress())
}

func (sm *Manager) SetLocalNodeAddress(addr *netip.AddrPort) {
	sm.Lock()
	defer sm.Unlock()
	sm.GetLocalNode().AddrPort = addr
}

func (sm *Manager) SendFromLocalNode(port uint16,
	b slice.Bytes) (e error) {
	
	sm.Lock()
	defer sm.Unlock()
	return sm.GetLocalNode().SendTo(port, b)
}

func (sm *Manager) ReceiveToLocalNode() <-chan slice.Bytes {
	sm.Lock()
	defer sm.Unlock()
	return sm.GetLocalNode().Transport.Receive()
}

func (sm *Manager) AddServiceToLocalNode(s *services.Service) (e error) {
	sm.Lock()
	defer sm.Unlock()
	return sm.GetLocalNode().AddService(s)
}

func (sm *Manager) GetLocalNodeRelayRate() (rate int) {
	sm.Lock()
	defer sm.Unlock()
	return sm.GetLocalNode().RelayRate
}

func (sm *Manager) GetLocalNodeIdentityBytes() (ident crypto.PubBytes) {
	sm.Lock()
	defer sm.Unlock()
	return sm.GetLocalNode().Identity.Bytes
}

func (sm *Manager) GetLocalNodeIdentityPrv() (ident *crypto.Prv) {
	sm.Lock()
	defer sm.Unlock()
	return sm.GetLocalNode().Identity.Prv
}

// SetLocalNode sets the engine's local Node.
func (sm *Manager) SetLocalNode(n *node.Node) {
	sm.Lock()
	defer sm.Unlock()
	sm.nodes[0] = n
}

// AddNodes adds a Node to a Nodes.
func (sm *Manager) AddNodes(nn ...*node.Node) {
	sm.Lock()
	defer sm.Unlock()
	sm.nodes = append(sm.nodes, nn...)
}
func (sm *Manager) FindNodeByIndex(i int) (no *node.Node) {
	sm.Lock()
	defer sm.Unlock()
	return sm.nodes[i]
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

// A SessionCache stores each of the 5 hops of a peer node.
type SessionCache map[nonce.ID]*sessions.Circuit

func (sm *Manager) UpdateSessionCache() {
	sm.Lock()
	defer sm.Unlock()
	// First we create SessionCache entries for all existing nodes.
	for i := range sm.nodes {
		_, exists := sm.SessionCache[sm.nodes[i].ID]
		if !exists {
			sm.SessionCache[sm.nodes[i].ID] = &sessions.Circuit{}
		}
	}
	// Place all sessions in their slots respective to their node.
	for _, v := range sm.Sessions {
		sm.SessionCache[v.Node.ID][v.Hop] = v
	}
}

func (sc SessionCache) Add(s *sessions.Data) SessionCache {
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

// PendingPayment accessors. For the same reason as the sessions, pending
// payments need to be accessed only with the node's mutex locked.

func (sm *Manager) AddPendingPayment(np *payments.Payment) {
	sm.Lock()
	defer sm.Unlock()
	log.D.F("%s adding pending payment %s for %v",
		sm.nodes[0].AddrPort.String(), np.ID,
		np.Amount)
	sm.PendingPayments = sm.PendingPayments.Add(np)
}
func (sm *Manager) DeletePendingPayment(preimage sha256.Hash) {
	sm.Lock()
	defer sm.Unlock()
	sm.PendingPayments = sm.PendingPayments.Delete(preimage)
}
func (sm *Manager) FindPendingPayment(id nonce.ID) (pp *payments.Payment) {
	sm.Lock()
	defer sm.Unlock()
	return sm.PendingPayments.Find(id)
}
func (sm *Manager) FindPendingPreimage(pi sha256.Hash) (pp *payments.Payment) {
	log.T.F("searching preimage %s", pi)
	sm.Lock()
	defer sm.Unlock()
	return sm.PendingPayments.FindPreimage(pi)
}
