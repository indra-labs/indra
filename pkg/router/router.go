package router

import "github.com/luke-park/ecdh25519"

type Router struct {
	priv *ecdh25519.PrivateKey
	*ecdh25519.PublicKey
}

func New() (r *Router, err error) {
	priv, err := ecdh25519.GenerateKey()
	if err != nil {
		return
	}
	r = &Router{priv: priv, PublicKey: priv.Public()}
	return
}
