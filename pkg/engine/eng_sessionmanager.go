package engine

import (
	"fmt"
	"net/netip"
	"sync"
	
	"git-indra.lan/indra-labs/lnd/lnd/lnwire"
	"github.com/gookit/color"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

type SessionManager struct {
	nodes           []*Node
	PendingPayments PendingPayments
	Sessions
	SessionCache
	sync.Mutex
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		SessionCache:    make(SessionCache),
		PendingPayments: make(PendingPayments, 0),
	}
}

// ClearPendingPayments is used only for debugging, removing all pending
// payments, making the engine forget about payments it received.
func (sm *SessionManager) ClearPendingPayments() {
	log.D.Ln("clearing pending payments")
	sm.PendingPayments = sm.PendingPayments[:0]
}

// ClearSessions is used only for debugging, removing all but the first session,
// which is the engine's initial return session.
func (sm *SessionManager) ClearSessions() {
	log.D.Ln("clearing sessions")
	sm.Sessions = sm.Sessions[:1]
}

func (sm *SessionManager) IncSession(id nonce.ID, msats lnwire.MilliSatoshi,
	sender bool, typ string) {
	
	sess := sm.FindSession(id)
	if sess != nil {
		sm.Lock()
		defer sm.Unlock()
		sess.IncSats(msats, sender, typ)
	}
}

func (sm *SessionManager) DecSession(id nonce.ID, msats int,
	sender bool, typ string) bool {
	
	sess := sm.FindSession(id)
	if sess != nil {
		sm.Lock()
		defer sm.Unlock()
		return sess.DecSats(lnwire.MilliSatoshi(msats/1024/1024),
			sender, typ)
	}
	return false
}

func (sm *SessionManager) GetNodeCircuit(id nonce.ID) (sce *Circuit,
	exists bool) {
	
	sm.Lock()
	defer sm.Unlock()
	sce, exists = sm.SessionCache[id]
	return
}

func (sm *SessionManager) AddSession(s *SessionData) {
	sm.Lock()
	defer sm.Unlock()
	// check for dupes
	for i := range sm.Sessions {
		if sm.Sessions[i].ID == s.ID {
			log.D.F("refusing to add duplicate session ID %x", s.ID)
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
func (sm *SessionManager) FindSession(id nonce.ID) *SessionData {
	sm.Lock()
	defer sm.Unlock()
	for i := range sm.Sessions {
		if sm.Sessions[i].ID == id {
			return sm.Sessions[i]
		}
	}
	return nil
}
func (sm *SessionManager) FindSessionByHeader(prvKey *prv.Key) *SessionData {
	sm.Lock()
	defer sm.Unlock()
	for i := range sm.Sessions {
		if sm.Sessions[i].Header.Prv.Key.Equals(&prvKey.Key) {
			return sm.Sessions[i]
		}
	}
	return nil
}
func (sm *SessionManager) FindSessionByHeaderPub(pubKey *pub.Key) *SessionData {
	sm.Lock()
	defer sm.Unlock()
	for i := range sm.Sessions {
		if sm.Sessions[i].Header.Pub.Equals(pubKey) {
			return sm.Sessions[i]
		}
	}
	return nil
}
func (sm *SessionManager) FindSessionPreimage(pr sha256.Hash) *SessionData {
	sm.Lock()
	defer sm.Unlock()
	for i := range sm.Sessions {
		if sm.Sessions[i].Preimage == pr {
			return sm.Sessions[i]
		}
	}
	return nil
}

func (sm *SessionManager) GetSessionsAtHop(hop byte) (s Sessions) {
	sm.Lock()
	defer sm.Unlock()
	for i := range sm.Sessions {
		if sm.Sessions[i].Hop == hop {
			s = append(s, sm.Sessions[i])
		}
	}
	return
}
func (sm *SessionManager) DeleteSession(id nonce.ID) {
	sm.Lock()
	defer sm.Unlock()
	for i := range sm.Sessions {
		if sm.Sessions[i].ID == id {
			// ProcessAndDelete from SessionData cache.
			sm.SessionCache[sm.Sessions[i].Node.ID][sm.Sessions[i].Hop] = nil
			// ProcessAndDelete from 
			sm.Sessions = append(sm.Sessions[:i], sm.Sessions[i+1:]...)
		}
	}
}

// IterateSessions calls a function for each entry in the Sessions slice.
//
// Do not call SessionManager methods within this function.
func (sm *SessionManager) IterateSessions(fn func(s *SessionData) bool) {
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
// Do not call SessionManager methods within this function.
func (sm *SessionManager) IterateSessionCache(fn func(n *Node,
	c *Circuit) bool) {
	
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

func (sm *SessionManager) GetSessionByIndex(i int) (s *SessionData) {
	sm.Lock()
	defer sm.Unlock()
	if len(sm.Sessions) > i {
		s = sm.Sessions[i]
	}
	return
}

func (sm *SessionManager) DeleteNodeAndSessions(id nonce.ID) {
	sm.Lock()
	defer sm.Unlock()
	var exists bool
	// If the node exists its ID is in the SessionCache.
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
	temp := make(Sessions, 0, len(sm.Sessions)-len(found))
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
func (sm *SessionManager) NodesLen() int {
	sm.Lock()
	defer sm.Unlock()
	return len(sm.nodes)
}

// GetLocalNode returns the engine's local Node.
func (sm *SessionManager) GetLocalNode() *Node { return sm.nodes[0] }

// GetLocalNodePaymentChan returns the engine's local Node PaymentChan.
func (sm *SessionManager) GetLocalNodePaymentChan() PaymentChan {
	return sm.nodes[0].PaymentChan
}

func (sm *SessionManager) GetLocalNodeAddress() (addr *netip.AddrPort) {
	// sm.Lock()
	// defer sm.Unlock()
	return sm.GetLocalNode().AddrPort
}

func (sm *SessionManager) GetLocalNodeAddressString() (s string) {
	return color.Yellow.Sprint(sm.GetLocalNodeAddress())
}

func (sm *SessionManager) SetLocalNodeAddress(addr *netip.AddrPort) {
	sm.Lock()
	defer sm.Unlock()
	sm.GetLocalNode().AddrPort = addr
}

func (sm *SessionManager) SendFromLocalNode(port uint16,
	b slice.Bytes) (e error) {
	
	sm.Lock()
	defer sm.Unlock()
	return sm.GetLocalNode().SendTo(port, b)
}

func (sm *SessionManager) ReceiveToLocalNode(port uint16) <-chan slice.Bytes {
	sm.Lock()
	defer sm.Unlock()
	if port == 0 {
		return sm.GetLocalNode().Transport.Receive()
	}
	return sm.GetLocalNode().ReceiveFrom(port)
}

func (sm *SessionManager) AddServiceToLocalNode(s *Service) (e error) {
	sm.Lock()
	defer sm.Unlock()
	return sm.GetLocalNode().AddService(s)
}

func (sm *SessionManager) GetLocalNodeRelayRate() (rate int) {
	sm.Lock()
	defer sm.Unlock()
	return sm.GetLocalNode().RelayRate
}

func (sm *SessionManager) GetLocalNodeIdentityBytes() (ident pub.Bytes) {
	sm.Lock()
	defer sm.Unlock()
	return sm.GetLocalNode().Identity.Bytes
}

func (sm *SessionManager) GetLocalNodeIdentityPrv() (ident *prv.Key) {
	sm.Lock()
	defer sm.Unlock()
	return sm.GetLocalNode().Identity.Prv
}

// SetLocalNode sets the engine's local Node.
func (sm *SessionManager) SetLocalNode(n *Node) {
	sm.Lock()
	defer sm.Unlock()
	sm.nodes[0] = n
}

// AddNodes adds a Node to a Nodes.
func (sm *SessionManager) AddNodes(nn ...*Node) {
	sm.Lock()
	defer sm.Unlock()
	sm.nodes = append(sm.nodes, nn...)
}
func (sm *SessionManager) FindNodeByIndex(i int) (no *Node) {
	sm.Lock()
	defer sm.Unlock()
	return sm.nodes[i]
}

// FindNodeByID searches for a Node by ID.
func (sm *SessionManager) FindNodeByID(i nonce.ID) (no *Node) {
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
func (sm *SessionManager) FindNodeByAddrPort(id *netip.AddrPort) (no *Node) {
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

// DeleteNodeByID deletes a node identified by an ID.
func (sm *SessionManager) DeleteNodeByID(ii nonce.ID) (e error) {
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
func (sm *SessionManager) DeleteNodeByAddrPort(ip *netip.AddrPort) (e error) {
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
// Do not call any SessionManager methods above inside this function or there
// will be a mutex double locking panic, except GetLocalNode.
func (sm *SessionManager) ForEachNode(fn func(n *Node) bool) {
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
type SessionCache map[nonce.ID]*Circuit

func (sm *SessionManager) UpdateSessionCache() {
	sm.Lock()
	defer sm.Unlock()
	// First we create SessionCache entries for all existing nodes.
	for i := range sm.nodes {
		_, exists := sm.SessionCache[sm.nodes[i].ID]
		if !exists {
			sm.SessionCache[sm.nodes[i].ID] = &Circuit{}
		}
	}
	// Place all sessions in their slots respective to their node.
	for _, v := range sm.Sessions {
		sm.SessionCache[v.Node.ID][v.Hop] = v
	}
}

func (sc SessionCache) Add(s *SessionData) SessionCache {
	var sce *Circuit
	var exists bool
	if sce, exists = sc[s.Node.ID]; !exists {
		sce = &Circuit{}
		sce[s.Hop] = s
		sc[s.Node.ID] = sce
		return sc
	}
	sc[s.Node.ID][s.Hop] = s
	return sc
}

// PendingPayment accessors. For the same reason as the sessions, pending
// payments need to be accessed only with the node's mutex locked.

func (sm *SessionManager) AddPendingPayment(np *Payment) {
	sm.Lock()
	defer sm.Unlock()
	log.D.F("%s adding pending payment %s for %v",
		sm.nodes[0].AddrPort.String(), np.ID,
		np.Amount)
	sm.PendingPayments = sm.PendingPayments.Add(np)
}
func (sm *SessionManager) DeletePendingPayment(preimage sha256.Hash) {
	sm.Lock()
	defer sm.Unlock()
	sm.PendingPayments = sm.PendingPayments.Delete(preimage)
}
func (sm *SessionManager) FindPendingPayment(id nonce.ID) (pp *Payment) {
	sm.Lock()
	defer sm.Unlock()
	return sm.PendingPayments.Find(id)
}
func (sm *SessionManager) FindPendingPreimage(pi sha256.Hash) (pp *Payment) {
	log.T.F("searching preimage %s", pi)
	sm.Lock()
	defer sm.Unlock()
	return sm.PendingPayments.FindPreimage(pi)
}
