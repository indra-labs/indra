package cryptorand

import (
	"math/rand"

	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

func Shuffle(l int, fn func(i, j int)) {
	rand.Seed(slice.GetCryptoRandSeed())
	rand.Shuffle(l, fn)
}
