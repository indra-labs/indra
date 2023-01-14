package response

import (
	"time"

	"github.com/indra-labs/indra/pkg/sha256"
)

type Hook struct {
	sha256.Hash
	Callback func()
	time.Time
}

type Hooks []Hook

func (h Hooks) Add(hash sha256.Hash, fn func()) (hh Hooks) {
	return append(h, Hook{Hash: hash, Callback: fn, Time: time.Now()})
}

func (h Hooks) Find(hash sha256.Hash) (hh Hooks) {
	for i := range h {
		if h[i].Hash == hash {
			h[i].Callback()
			hh = append(h[:i], h[i+1:]...)
			return
		}
	}
	return
}
