package relay

// SelectUnusedCircuit accepts an array of 5 Node entries where all or some are
// empty and picks nodes for the remainder that do not have a hop at that
// position.
func (sm *SessionManager) SelectUnusedCircuit() (c [5]*Node) {
	sm.Lock()
	defer sm.Unlock()
	// Create a shuffled slice of Nodes to randomise the selection process.
	nodeList := make([]*Node, len(sm.nodes)-1)
	copy(nodeList, sm.nodes[1:])
	for i := range nodeList {
		if _, ok := sm.SessionCache[nodeList[i].ID]; !ok {
			log.T.F("adding session cache entry for node %s", nodeList[i].ID)
			sm.SessionCache[nodeList[i].ID] = &Circuit{}
		}
	}
	var counter int
out:
	for counter < 5 {
		for i := range sm.SessionCache {
			if counter == 5 {
				break out
			}
			if sm.SessionCache[i][counter] == nil {
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
