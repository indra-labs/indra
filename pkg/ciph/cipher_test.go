package ciph

import (
	"crypto/aes"
	"crypto/cipher"
	"testing"

	"github.com/Indra-Labs/indra/pkg/key/ecdh"
	"github.com/Indra-Labs/indra/pkg/key/prv"
	"github.com/Indra-Labs/indra/pkg/key/pub"
	nonce2 "github.com/Indra-Labs/indra/pkg/nonce"
	"github.com/Indra-Labs/indra/pkg/sha256"
	"github.com/Indra-Labs/indra/pkg/testutils"
)

func TestGenerateCompositeCipher(t *testing.T) {
	var cifkeys []*CipherKeys
	var e error
	for i := 0; i < 3; i++ {
		var p1, p2 *prv.Key
		if p1, e = prv.GenerateKey(); check(e) {
			t.Error(e)
			t.FailNow()
		}
		pub1 := pub.Derive(p1)
		if p2, e = prv.GenerateKey(); check(e) {
			t.Error(e)
			t.FailNow()
		}
		cifkeys = append(cifkeys, &CipherKeys{p2, pub1})
	}
	var m, original []byte
	m, _, e = testutils.GenerateTestMessage(1024)
	log.I.S(sha256.Single(m))
	original = make([]byte, len(m))
	copy(original, m)
	var compositeSecret []byte
	if compositeSecret, e =
		GenerateCompositeCipher(cifkeys...); check(e) {

		t.Error(e)
		t.FailNow()
	}
	var block cipher.Block
	if block, e = aes.NewCipher(compositeSecret); check(e) {
		t.Error(e)
		t.FailNow()
	}
	nonce := nonce2.New()
	Encipher(block, nonce, m)
	log.I.S(sha256.Single(m))
	// for i := range cifkeys {
	// 	var blk cipher.Block
	// 	if blk, e = GetBlock(cifkeys[i].From,
	// 		cifkeys[i].To); check(e) {
	// 		t.Error(e)
	// 		t.FailNow()
	// 	}
	// 	Encipher(blk, nonce, m)
	// }
	// log.I.S(sha256.Single(m), sha256.Single(original))
	for i := range cifkeys {
		var blk cipher.Block
		if blk, e = GetBlock(cifkeys[i].From,
			cifkeys[i].To); check(e) {
			t.Error(e)
			t.FailNow()
		}
		Encipher(blk, nonce, original)
	}
	log.I.S(sha256.Single(original))

}
func TestCompositing(t *testing.T) {
	var cifkeys []*CipherKeys
	var e error
	for i := 0; i < 3; i++ {
		var p1, p2 *prv.Key
		if p1, e = prv.GenerateKey(); check(e) {
			t.Error(e)
			t.FailNow()
		}
		pub1 := pub.Derive(p1)
		if p2, e = prv.GenerateKey(); check(e) {
			t.Error(e)
			t.FailNow()
		}
		cifkeys = append(cifkeys, &CipherKeys{p2, pub1})
	}
	var secrets [][]byte
	for i := range cifkeys {
		secrets = append(secrets, ecdh.Compute(cifkeys[i].
			From, cifkeys[i].To))
	}
	log.I.S(secrets)
	var cs []byte
	if cs, e = CombineCiphers(secrets...); check(e) {
		t.Error(e)
		t.FailNow()
	}
	log.I.S(cs)

	// this proves we can retreive the third cipher if we know
	// the first two
	for _, i := range secrets[:2] {
		for j := range cs {
			cs[j] ^= i[j]
		}
	}
	log.I.S(cs, secrets[2])

	testWords := []byte("this is a test")
	testBytes1 := make([]byte, len(testWords))
	copy(testBytes1, testWords)
	testBytes2 := make([]byte, len(testWords))
	copy(testBytes2, testWords)
	log.I.S(testWords, testBytes1, testBytes2)

	// regenerate the combined cipher
	if cs, e = CombineCiphers(secrets...); check(e) {
		t.Error(e)
		t.FailNow()
	}
	var nonces []nonce2.IV
	for i := 0; i < 3; i++ {
		nonces = append(nonces, nonce2.New())
	}
	log.I.S(nonces)
	n := nonce2.New()
	tmp := make(nonce2.IV, nonce2.IVLen)
	copy(tmp, nonces[0])
	nt := touint64(tmp)
	for i := range nonces {
		if i != 0 {
			u64 := touint64(nonces[i])
			xor64(nt, u64)
		}
	}
	n = tobytes(nt)
	log.I.S(n)
	var cb cipher.Block
	if cb, e = aes.NewCipher(cs); !check(e) {
	}
	Encipher(cb, n, testBytes1)
	log.I.S(testBytes1)
	for i := range cifkeys {
		var blk cipher.Block
		if blk, e = GetBlock(cifkeys[i].From,
			cifkeys[i].To); check(e) {
			t.Error(e)
			t.FailNow()
		}
		Encipher(blk, nonces[i], testBytes2)
	}
	log.I.S(testBytes2)
}
