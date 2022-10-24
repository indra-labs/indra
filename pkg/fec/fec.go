package fec

import (
	"github.com/Indra-Labs/indra"
	log2 "github.com/cybriq/proc/pkg/log"
	"github.com/templexxx/reedsolomon"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

// Encode ...
func Encode(data [][]byte, d, p int) (chunks [][]byte, e error) {
	dl := len(data)
	pl := dl * p / d
	var rs *reedsolomon.RS
	if rs, e = reedsolomon.New(dl, pl); check(e) {
		return
	}
	_ = rs
	return
}

// Decode takes a set of shards and if there is sufficient to reassemble,
// returns the corrected data
func Decode(chunks [][]byte) (data []byte, e error) {

	return
}
