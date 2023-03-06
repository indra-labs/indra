package engine

import (
	"fmt"
	
	"git-indra.lan/indra-labs/indra"
	"git-indra.lan/indra-labs/indra/pkg/crypto/ciph"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/util/octet"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

const (
	MagicLen    = 2
	ErrTooShort = "'%s' message  minimum size: %d got: %d"
)

func TooShort(got, found int, magic string) (e error) {
	if got >= found {
		return
	}
	e = fmt.Errorf(ErrTooShort, magic, got, found)
	return
	
}

func BudgeUp(s *octet.Splice) (o *octet.Splice) {
	o = s
	start := o.GetCursor()
	copy(o.GetRange(-1, -1), s.GetRange(start, -1))
	copy(s.GetRange(o.Len()-start, -1), slice.NoisePad(start))
	return
}

func FormatReply(header, res slice.Bytes, ciphers [3]sha256.Hash,
	nonces [3]nonce.IV) (rb *octet.Splice) {
	
	rb = octet.New(ReverseHeaderLen + len(res))
	check(rb.Bytes(header[:ReverseHeaderLen]).Bytes(res))
	start := rb.GetCursor()
	for i := range ciphers {
		blk := ciph.BlockFromHash(ciphers[i])
		ciph.Encipher(blk, nonces[2-i], rb.GetRange(start, -1))
	}
	return
}
