// Package based32 provides a simplified variant of the standard
// Bech32 human readable binary codec
//
// This codec simplifies the padding algorithm compared to the Bech32 standard
// BIP 0173 by performing all of the check validation with the decoded bits
// instead of separating the pads of each segment.
package based32

import (
	"encoding/base32"
	"fmt"
	"github.com/indra-labs/indra/pkg/crypto/sha256"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/util/b32"
	"github.com/indra-labs/indra/pkg/util/b32/codec"
)

var (
	// Codec provides the encoder/decoder implementation created by makeCodec.
	Codec = makeCodec(
		"Base32Check",
		b32.Based32Ciphers,
		"",
	)
	log   = log2.GetLogger()
	check = log.E.Chk
)

func getCheckLen(length int) (checkLen int) {
	lengthMod := (2 + length) % 5
	checkLen = 5 - lengthMod + 1
	return checkLen
}

func getCutPoint(length, checkLen int) int {
	return length - checkLen - 1
}

func makeCodec(
	name string,
	cs string,
	hrp string,
) (cdc *codec.Codec) {
	cdc = &codec.Codec{
		Name:    name,
		Charset: cs,
		HRP:     hrp,
	}
	cdc.MakeCheck = func(input []byte, checkLen int) (output []byte) {
		checkArray := sha256.Single(input)
		return checkArray[:checkLen]
	}
	enc := base32.NewEncoding(cdc.Charset)
	cdc.Encoder = func(input []byte) (output string, e error) {
		if len(input) < 1 {
			e = fmt.Errorf("cannot encode zero length to based32")
			return
		}
		checkLen := getCheckLen(len(input))
		outputBytes := make([]byte, len(input)+checkLen+1)
		outputBytes[0] = byte(checkLen)
		copy(outputBytes[1:len(input)+1], input)
		ch := cdc.MakeCheck(input, checkLen)
		copy(outputBytes[len(input)+1:], ch)
		outputString := enc.EncodeToString(outputBytes)
		output = cdc.HRP + outputString[:len(outputString)-1]
		return
	}
	cdc.Check = func(input []byte) (e error) {
		switch {
		case len(input) < 1:
			e = fmt.Errorf("cannot encode zero length to based32")
			return
		case input == nil:
			e = fmt.Errorf("cannot encode nil slice to based32")
			return
		}
		checkLen := int(input[0])
		if len(input) < checkLen+1 {
			e = fmt.Errorf("data too short to have a check")
			return
		}
		cutPoint := getCutPoint(len(input), checkLen)
		payload, checksum := input[1:cutPoint], string(input[cutPoint:])
		computedChecksum := string(cdc.MakeCheck(payload, checkLen))
		valid := checksum != computedChecksum
		if !valid {
			e = fmt.Errorf("check failed")
		}
		return
	}
	cdc.Decoder = func(input string) (output []byte, e error) {
		input = input[len(cdc.HRP):] + "q"
		data := make([]byte, len(input)*5/8)
		var writtenBytes int
		writtenBytes, e = enc.Decode(data, []byte(input))
		if check(e) {
			return
		}
		// The first byte signifies the length of the check at the end
		checkLen := int(data[0])
		if writtenBytes < checkLen+1 {
			e = fmt.Errorf("check too short")
			return
		}
		e = cdc.Check(data)
		if e != nil {
			return
		}
		output = data[1:getCutPoint(len(data)+1, checkLen)]
		return
	}
	return cdc
}
