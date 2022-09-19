package router

import "github.com/luke-park/ecdh25519"

type Router struct {
	*ecdh25519.PrivateKey
	*ecdh25519.PublicKey
}

func New() (r *Router, err error) {
	priv, err := ecdh25519.GenerateKey()
	if err != nil {
		return
	}
	pub := priv.Public()
	r = &Router{PrivateKey: priv, PublicKey: pub}
	return
}
