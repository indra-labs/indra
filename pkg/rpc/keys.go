package rpc

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"github.com/btcsuite/btcd/btcutil/base58"
	"golang.org/x/crypto/curve25519"
	"golang.zx2c4.com/wireguard/device"
)

const (
	RPCPrivateKeySize = 32
	RPCPublicKeySize  = 32
)

type (
	RPCPrivateKey [RPCPrivateKeySize]byte
	RPCPublicKey  [RPCPublicKeySize]byte
)

var (
	DefaultRPCPrivateKey RPCPrivateKey
	DefaultRPCPublicKey  RPCPublicKey
)

func NewPrivateKey() (sk RPCPrivateKey, err error) {

	_, err = rand.Read(sk[:])

	sk[0] &= 248
	sk[31] = (sk[31] & 127) | 64

	return
}

func (key RPCPrivateKey) Equals(tar RPCPrivateKey) bool {
	return subtle.ConstantTimeCompare(key[:], tar[:]) == 1
}

func (sk RPCPrivateKey) IsZero() bool {
	var zero RPCPrivateKey
	return sk.Equals(zero)
}

func (sk *RPCPrivateKey) PubKey() (pk RPCPublicKey) {
	apk := (*[device.NoisePublicKeySize]byte)(&pk)
	ask := (*[RPCPrivateKeySize]byte)(sk)
	curve25519.ScalarBaseMult(apk, ask)

	return
}

func (sk *RPCPrivateKey) AsDeviceKey() device.NoisePrivateKey {
	return device.NoisePrivateKey(*sk)
}

func (sk RPCPrivateKey) Encode() (key string) {

	key = base58.Encode(sk[:])

	return
}

func (sk RPCPrivateKey) HexString() string {
	return hex.EncodeToString(sk[:])
}

func (sk *RPCPrivateKey) Decode(key string) {
	copy(sk[:], base58.Decode(key))
}

func DecodePrivateKey(key string) RPCPrivateKey {

	var pk RPCPrivateKey

	pk.Decode(key)

	return pk
}

func (sk RPCPublicKey) AsDeviceKey() device.NoisePublicKey {
	return device.NoisePublicKey(sk)
}

func (sk RPCPublicKey) Encode() (key string) {

	key = base58.Encode(sk[:])

	return
}

func (sk RPCPublicKey) HexString() string {
	return hex.EncodeToString(sk[:])
}

func (sk *RPCPublicKey) Decode(key string) {
	copy(sk[:], base58.Decode(key))
}

func DecodePublicKey(key string) *RPCPublicKey {

	var pk RPCPublicKey

	pk.Decode(key)

	return &pk
}
