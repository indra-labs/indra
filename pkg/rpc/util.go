package rpc

import (
	"math/rand"
	"time"
)

func genRandomPort(offset int) uint16 {

	rand.Seed(time.Now().Unix())

	return uint16(rand.Intn(65534-offset) + offset)
}
