package sifr

//
// import (
// 	"crypto/aes"
// 	"crypto/cipher"
// 	"crypto/rand"
//
// 	"github.com/Indra-Labs/indra/pkg/schnorr"
// )
//
// type Codec struct {
// 	priv    []*schnorr.Privkey
// 	pub     []*schnorr.Pubkey
// 	gcm     []cipher.AEAD
// 	cursor  int
// 	buffers int
// }
//
// // New creates a new Codec with a number of buffers for old ciphers
// func New(bufs int, recvPubKey *schnorr.PubkeyBytes) (cdc *Codec, e error) {
// 	cdc = &Codec{
// 		priv:    make([]*schnorr.Privkey, bufs),
// 		pub:     make([]*schnorr.Pubkey, bufs),
// 		gcm:     make([]cipher.AEAD, bufs),
// 		cursor:  0,
// 		buffers: bufs,
// 	}
// 	if cdc.priv[cdc.cursor], e = schnorr.GeneratePrivkey(); log.E.Chk(e) {
// 		return
// 	}
// 	if cdc.pub[cdc.cursor], e = recvPubKey.Deserialize(); log.E.Chk(e) {
// 		return
// 	}
// 	secret := cdc.priv[cdc.cursor].ECDH(cdc.pub[cdc.cursor])
// 	var ci cipher.Block
// 	if ci, e = aes.NewCipher(secret); log.E.Chk(e) {
// 		return
// 	}
// 	if cdc.gcm[cdc.cursor], e = cipher.NewGCM(ci); log.E.Chk(e) {
// 	}
// 	return
// }
//
// func (c *Codec) advanceCursor() {
// 	c.cursor++
// 	if c.cursor >= c.buffers {
// 		c.cursor = 0
// 	}
// 	return
// }
//
// func (c *Codec) Send() (e error) {
//
// 	return
// }
//
// func (c *Codec) Reply() (e error) {
//
// 	return
// }
//
// const NonceLen = 12
//
// type Nonce [NonceLen]byte
//
// // GetNonce reads from a cryptographically secure random number source
// func GetNonce() (nonce *Nonce) {
// 	if _, e := rand.Read(nonce[:]); log.E.Chk(e) {
// 	}
// 	return
// }
