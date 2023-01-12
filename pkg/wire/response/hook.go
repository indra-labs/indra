package response

import (
	"time"

	"github.com/indra-labs/indra/pkg/sha256"
	"github.com/indra-labs/indra/pkg/slice"
)

type Hook struct {
	sha256.Hash
	Callback func(b slice.Bytes)
	time.Time
}

type Hooks []Hook

func (h Hooks) Add(hash sha256.Hash, fn func(b slice.Bytes)) (hh Hooks) {
	return append(h, Hook{Hash: hash, Callback: fn, Time: time.Now()})
}

func (h Hooks) Find(hash sha256.Hash, b slice.Bytes) (hh Hooks) {
	for i := range h {
		if h[i].Hash == hash {
			h[i].Callback(b)
			hh = append(h[:i], h[i+1:]...)
			return
		}
	}
	return
}
