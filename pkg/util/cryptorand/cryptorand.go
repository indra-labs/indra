// Package cryptorand augments the standard math/rand library with cryptographic entropy seeding.
package cryptorand

import (
	rand2 "crypto/rand"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/util/slice"
	"math/rand"
)

var (
	log   = log2.GetLogger()
	check = log.E.Chk
)

func GetSeed() int64 {
	rBytes := make([]byte, 8)
	if n, e := rand2.Read(rBytes); n != 8 && check(e) {
		return 0
	}
	return int64(slice.DecodeUint64(rBytes))
}

func IntN(n int) int {
	rand.Seed(GetSeed())
	return rand.Intn(n)
}

func Shuffle(l int, fn func(i, j int)) {
	rand.Seed(GetSeed())
	rand.Shuffle(l, fn)
}
