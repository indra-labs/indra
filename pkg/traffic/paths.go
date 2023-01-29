package traffic

import (
	"math/rand"

	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

func (pm *Payments) Select(hops []byte, alreadyHave Sessions) (so Sessions) {
	pm.Lock()
	defer pm.Unlock()
	ws := make(Sessions, 0)
	// todo: later on we want to pre-thin this according to configuration.
out:
	for i := range pm.Sessions {
		if pm.Sessions[i] == nil {
			log.D.Ln("nil session", i)
			continue
		}
		for j := range alreadyHave {
			// We won't select any given in the alreadyHave list
			if alreadyHave[j] == nil {
				continue
			}
			if pm.Sessions[i].ID == alreadyHave[j].ID {
				continue out
			}
		}
		ws = append(ws, pm.Sessions[i])
	}
	// Shuffle the copy of the sessions.
	rand.Seed(slice.GetCryptoRandSeed())
	rand.Shuffle(len(ws), func(i, j int) {
		ws[i], ws[j] = ws[j], ws[i]
	})
	// log.T.S(ws)
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
	return
}
