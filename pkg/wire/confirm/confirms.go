package confirm

import (
	"sync"
	"time"

	"github.com/indra-labs/indra/pkg/nonce"
)

type Hook func(cf nonce.ID)

type Callback struct {
	nonce.ID
	time.Time
	Onion *OnionSkin
	Hook  Hook
}

type Confirms struct {
	sync.Mutex
	Cnf []Callback
}

func NewConfirms() *Confirms {
	cn := Confirms{
		Cnf: make([]Callback, 0),
	}
	return &cn
}

func (cn *Confirms) Add(cb *Callback) {
	cn.Lock()
	(*cn).Cnf = append((*cn).Cnf, *cb)
	cn.Unlock()
}

func (cn *Confirms) Confirm(id nonce.ID) bool {
	cn.Lock()
	defer cn.Unlock()
	for i := range (*cn).Cnf {
		if id == (*cn).Cnf[i].ID {
			(*cn).Cnf[i].Hook(id)
			// delete Callback.
			end := i + 1
			// if this is the last one, the end is the last one
			// also.
			if end > len((*cn).Cnf)-1 {
				end = len((*cn).Cnf) - 1
			}
			(*cn).Cnf = append((*cn).Cnf[:i], (*cn).Cnf[end:]...)
			return true
		}
	}
	return false
}

// Flush clears out entries older than timePast.
func (cn *Confirms) Flush(timePast time.Time) {
	cn.Lock()
	defer cn.Unlock()
	var foundCount int
	for i := range (*cn).Cnf {
		if (*cn).Cnf[i].Time.Before(timePast) {
			foundCount++
		}
	}
	if foundCount > 0 {
		cnNew := NewConfirms()
		for i := range (*cn).Cnf {
			if !(*cn).Cnf[i].Time.Before(timePast) {
				(*cnNew).Cnf = append((*cnNew).Cnf,
					(*cn).Cnf[i])
			}
		}
		(*cn).Cnf = (*cnNew).Cnf
	}
}
