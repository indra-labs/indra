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
	PrivateKeySize = 32
	PublicKeySize  = 32
)

type (
	PrivateKey [PrivateKeySize]byte
	PublicKey  [PublicKeySize]byte
)

func NewPrivateKey() (*PrivateKey, error) {

	var err error
	var sk PrivateKey

	_, err = rand.Read(sk[:])

	sk[0] &= 248
	sk[31] = (sk[31] & 127) | 64

	return &sk, err
}

func (key PrivateKey) Equals(tar PrivateKey) bool {
	return subtle.ConstantTimeCompare(key[:], tar[:]) == 1
}

func (sk PrivateKey) IsZero() bool {
	var zero PrivateKey
	return sk.Equals(zero)
}

func (sk *PrivateKey) PubKey() (pk PublicKey) {
	apk := (*[device.NoisePublicKeySize]byte)(&pk)
	ask := (*[PrivateKeySize]byte)(sk)
	curve25519.ScalarBaseMult(apk, ask)

	return
}

func (sk *PrivateKey) AsDeviceKey() device.NoisePrivateKey {
	return device.NoisePrivateKey(*sk)
}

func (sk PrivateKey) Encode() (key string) {

	key = base58.Encode(sk[:])

	return
}

func (sk PrivateKey) Bytes() []byte {
	return sk[:]
}

func (sk PrivateKey) HexString() string {
	return hex.EncodeToString(sk[:])
}

func (sk *PrivateKey) Decode(key string) {
	copy(sk[:], base58.Decode(key))
}

func (sk *PrivateKey) DecodeBytes(key []byte) {
	copy(sk[:], key)
}

func DecodePrivateKey(key string) PrivateKey {

	var pk PrivateKey

	pk.Decode(key)

	return pk
}

func (sk PublicKey) AsDeviceKey() device.NoisePublicKey {
	return device.NoisePublicKey(sk)
}

func (sk PublicKey) Encode() (key string) {

	key = base58.Encode(sk[:])

	return
}

func (sk PublicKey) HexString() string {
	return hex.EncodeToString(sk[:])
}

func (sk *PublicKey) Decode(key string) {
	copy(sk[:], base58.Decode(key))
}

func DecodePublicKey(key string) PublicKey {

	var pk PublicKey

	pk.Decode(key)

	return pk
}
