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
	
	"git-indra.lan/indra-labs/indra"
	"git-indra.lan/indra-labs/indra/pkg/b32/codec"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

// Charset is the set of characters used in the data section of bech32 strings.
// Note that this is ordered, such that for a given Charset[i], i is the binary
// value of the character.
const Charset = "abcdefghijklmnopqrstuvwxyz234679"

// Codec provides the encoder/decoder implementation created by makeCodec.
var Codec = makeCodec(
	"Base32Check",
	Charset,
	"",
)

func getCheckLen(length int) (checkLen int) {
	lengthMod := (2 + length) % 5
	checkLen = 5 - lengthMod + 1
	return checkLen
}

func getCutPoint(length, checkLen int) int {
	return length - checkLen - 1
}

// Shift5bitsLeft allows the elimination of the first 5 bits of the value,
// which are always zero in standard base32 encoding when encoding a public key.
func Shift5bitsLeft(b slice.Bytes) (o slice.Bytes) {
	o = make(slice.Bytes, len(b))
	for i := range b {
		if i != len(b)-1 {
			o[i] = b[i] << 5
			o[i] += b[i+1] >> 3
		} else {
			o[i] = b[i] << 5
		}
	}
	return
}
func Shift5bitsRight(b slice.Bytes) (o slice.Bytes) {
	o = make(slice.Bytes, len(b))
	for i := range b {
		if i == 0 {
			o[i] = b[i] >> 5
		} else {
			o[i] = b[i] >> 5
			o[i] += b[i-1] << 3
		}
	}
	return
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
		copy(outputBytes[len(input)+1:], cdc.MakeCheck(input, checkLen))
		outputBytes = Shift5bitsLeft(outputBytes)
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
		data = Shift5bitsRight(data)
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
