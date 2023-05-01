package holo

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"
	
	"github.com/jbarham/primegen"
	
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
)

func DecodeUint64(b []byte) uint64 { return get64(b) }

var put64 = binary.LittleEndian.PutUint64
var get64 = binary.LittleEndian.Uint64

func Shuffle(l int, fn func(i, j int)) {
	rand.Seed(time.Now().Unix())
	rand.Shuffle(l, fn)
}

func TestNew(t *testing.T) {
	log2.SetLogLevel(log2.Trace)
	primer := primegen.New()
	counter := 0
	var prime uint8 = 1
	var primes []uint8
	// for i := 2; prime < 72; i++ {
	for i := 2; len(primes) < 10; i++ {
		prime = uint8(primer.Next())
		primes = append(primes, prime)
		counter++
	}
	fmt.Println(primes)
	fmt.Printf("\nprime factors addded\n\n")
	fmt.Printf(
		"|index|prime|cumulative product|binary|product |64 bit prod |64 bit" +
			"  binary prod               | and" +
			"     |    |                                  |\n")
	fmt.Printf(
		"|-----|-----|------------------|------|--------|------------|----------------------------------|---------|----|----------------------------------|\n")
	prod := primes[0]
	prod6 := uint64(primes[0])
	var prevs []uint8
	for i := range primes {
		prevs = append(prevs, prod)
		prod *= primes[i]
		prod6 *= uint64(primes[i])
		fmt.Printf("|%5d|%5d|%18d|%6s|%8s|%12d|%34s|%9s|%4d|%34s|\n",
			i,
			primes[i],
			prod,
			strconv.FormatInt(int64(primes[i]), 2),
			strconv.FormatInt(int64(prod), 2),
			prod6,
			strconv.FormatInt(int64(prod6), 2),
			strconv.FormatInt(int64(uint64(prod)&prod6), 2),
			uint64(prod)&prod6,
			strconv.FormatInt(int64(uint64(prod)|prod6), 2),
		)
	}
	prevs = append(prevs, prod)
	for i := range primes {
		if i >= len(primes)/2 {
			continue
		}
		primes[i], primes[len(primes)-1-i] = primes[len(primes)-1-i], primes[i]
		prevs[i], prevs[len(prevs)-1-i] = prevs[len(prevs)-1-i],
			prevs[i]
	}
	
	// fmt.Println(prevs)
	
	// prevs = prevs[1:]
	last := prevs[0]
	// fmt.Println(prevs)
	fmt.Printf("\nprime factors removed\n\n")
	fmt.Printf("|index|prime|  cumulative product|\n")
	fmt.Printf("|-----|-----|--------------------|\n")
	var m, n uint8
	_ = n
	for i := range primes {
		if i == 0 {
			continue
		}
		
		fmt.Printf("|%5d|%5d|%20d|%20d\n",
			len(primes)-i-1, primes[i], last, m)
		
		if m != 0 {
			// log.D.Ln("remainder not zero!")
			// return
		}
		if m == 0 {
		}
	}
}
