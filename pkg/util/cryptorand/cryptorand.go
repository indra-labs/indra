package cryptorand

import (
	rand2 "crypto/rand"
	"math/rand"

	"git-indra.lan/indra-labs/indra"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

func Shuffle(l int, fn func(i, j int)) {
	rand.Seed(GetSeed())
	rand.Shuffle(l, fn)
}

func GetSeed() int64 {
	rBytes := make([]byte, 8)
	if n, e := rand2.Read(rBytes); n != 8 && check(e) {
		return 0
	}
	return int64(slice.DecodeUint64(rBytes))
}
