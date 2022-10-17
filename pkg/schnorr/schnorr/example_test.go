// Copyright (c) 2020-2021 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package schnorr_test

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/Indra-Labs/indra/pkg/schnorr/schnorr"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// This example demonstrates signing a message with the EC-Schnorr-INDRAv0 scheme
// using a secp256k1 private key that is first parsed from raw bytes and
// serializing the generated signature.
func ExampleSign() {
	// Decode a hex-encoded private key.
	pkBytes, err := hex.DecodeString("22a47fa09a223f2aa079edf85a7c2d4f8720ee6" +
		"3e502ee2869afab7de234b80c")
	if err != nil {
		fmt.Println(err)
		return
	}
	privKey := secp256k1.PrivKeyFromBytes(pkBytes)

	// Sign a message using the private key.
	message := "test message"
	messageHash := sha256.Sum256([]byte(message))
	signature, err := schnorr.Sign(privKey, messageHash[:])
	if err != nil {
		fmt.Println(err)
		return
	}

	// Serialize and display the signature.
	fmt.Printf("Serialized Signature: %x\n", signature.Serialize())

	// Verify the signature for the message using the public key.
	pubKey := privKey.PubKey()
	verified := signature.Verify(messageHash[:], pubKey)
	fmt.Printf("Signature Verified? %v\n", verified)

	// Output:
	// Serialized Signature: 56775710b6270bc34e4f364f10cd2583cbb8a2c4c3b578404d84e00f2f174e342f4eb1da7c34cccfbdd5b123157bc3f1d20ecbee8f42f05339f9a8814b179a1e
	// Signature Verified? true
}

// Currently disabled due to change of hash function
//
// // This example demonstrates verifying an EC-Schnorr-INDRAv0 signature against a
// // public key that is first parsed from raw bytes.  The signature is also parsed
// // from raw bytes.
// func ExampleSignature_Verify() {
// 	// Decode hex-encoded serialized public key.
// 	pubKeyBytes, err := hex.DecodeString("02a673638cb9587cb68ea08dbef685c6f2d" +
// 		"2a751a8b3c6f2a7e9a4999e6e4bfaf5")
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}
// 	pubKey, err := schnorr.ParsePubKey(pubKeyBytes)
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}
//
// 	// Decode hex-encoded serialized signature.
// 	sigBytes, err := hex.DecodeString("970603d8ccd2475b1ff66cfb3ce7e622c59383" +
// 		"48304c5a7bc2e6015fb98e3b457d4e912fcca6ca87c04390aa5e6e0e613bbbba7ffd" +
// 		"6f15bc59f95bbd92ba50f0")
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}
// 	signature, err := schnorr.ParseSignature(sigBytes)
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}
//
// 	// Verify the signature for the message using the public key.
// 	message := "test message"
// 	messageHash := sha256.Sum256([]byte(message))
// 	verified := signature.Verify(messageHash[:], pubKey)
// 	fmt.Println("Signature Verified?", verified)
//
// 	// Output:
// 	// Signature Verified? true
// }
