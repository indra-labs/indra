package sess

import (
	"git.indra-labs.org/dev/ind/pkg/engine/node"
	"git.indra-labs.org/dev/ind/pkg/engine/sessions"
	"git.indra-labs.org/dev/ind/pkg/util/cryptorand"
)

// SelectHops picks out a set of sessions to use in a circuit.
func (sm *Manager) SelectHops(hops []byte, alreadyHave sessions.Sessions,
	note string) (so sessions.Sessions) {
	log.T.Ln(sm.GetLocalNodeAddressString(), "selecting hops", note)
	sm.Lock()
	defer sm.Unlock()
	ws := make(sessions.Sessions, 0)
out:
	for i := range sm.Sessions {
		if sm.Sessions[i] == nil {
			log.D.Ln("nil session", i)
			continue
		}
		for j := range alreadyHave {
			// We won't select any given in the alreadyHave list
			if alreadyHave[j] == nil {
				continue
			}
			if sm.Sessions[i].Header.Bytes == alreadyHave[j].Header.Bytes {
				continue out
			}
		}
		ws = append(ws, sm.Sessions[i])
	}
	// Shuffle the copy of the candidates.
	cryptorand.Shuffle(len(ws), func(i, j int) {
		ws[i], ws[j] = ws[j], ws[i]
	})
	log.T.Ln("shuffled", len(ws), "candidate sessions")
	// Iterate the available sessions picking the first matching hop, then
	// prune it from the temporary slice and advance the cursor, wrapping
	// around at end.
	if len(alreadyHave) != len(hops) {
		log.E.Ln("selection requires equal length of hops and sessions")
		return
	}
	so = alreadyHave
	for i := range hops {
		var cur int
	out2:
		for {
			if ws[cur] == nil {
				cur++
				continue
			}
			if so[i] != nil {
				break out2
			}
			if ws[cur].Hop == hops[i] {
				for _, v := range so[:i] {
					for a := range v.Node.Addresses {
						if v.Node.Addresses[a].String() == sm.
							GetLocalNodeAddressString() ||
							v.Node.Addresses[a].String() == ws[cur].Node.
								Addresses[a].String() {
							continue
						}

					}
				}
				so[i] = ws[cur]
				ws[cur] = nil
				cur = 0
				break
			}
			cur++
			if cur == len(ws) {
				cur = 0
			}
		}
	}
	var str string
	for i := range so {
		for j := range so[i].Node.Addresses {
			str += so[i].Node.Addresses[j].String() + " "
		}
	}
	log.D.F("circuit\n%s", str)
	return
}

// SelectUnusedCircuit accepts an array of 5 Node entries where all or some are
// empty and picks nodes for the remainder that do not have a hop at that
// position.
func (sm *Manager) SelectUnusedCircuit() (c [5]*node.Node) {
	sm.Lock()
	defer sm.Unlock()
	// Create a shuffled slice of Nodes to randomise the selection process.
	nodeList := make([]*node.Node, len(sm.nodes)-1)
	copy(nodeList, sm.nodes[1:])
	for i := range nodeList {
		if _, ok := sm.CircuitCache[nodeList[i].ID]; !ok {
			log.T.F("adding session cache entry for node %s", nodeList[i].ID)
			sm.CircuitCache[nodeList[i].ID] = &sessions.Circuit{}
		}
	}
	var counter int
out:
	for counter < 5 {
		for i := range sm.CircuitCache {
			if counter == 5 {
				break out
			}
			if sm.CircuitCache[i][counter] == nil {
				for j := range nodeList {
					if nodeList[j].ID == i {
						c[counter] = nodeList[j]
						counter++
						break
					}
				}
			}
		}
	}
	return
}

// StandardCircuit is a slice defining a standard circuit (5 is the return session).
func StandardCircuit() []byte { return []byte{0, 1, 2, 3, 4, 5} }
