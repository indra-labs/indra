package rpc

func configureTunnel() {

	if !o.tunEnable {
		return
	}

	log.I.Ln("enabling rpc tunnel")

	configureTunnelPort()

	log.I.Ln("rpc tunnel listeners:")
	log.I.F("- [/ip4/0.0.0.0/udp/%d /ip6/:::/udp/%d]", o.tunPort, o.tunPort)

	configureTunnelKey()
	configurePeerWhitelist()
}

func configureTunnelKey() {

	log.I.Ln("looking for key in storage")

	var err error

	tunKey, err = o.store.GetKey()

	if err == nil {

		log.I.Ln("rpc tunnel public key:")
		log.I.Ln("-", tunKey.PubKey().Encode())

		return
	}

	if err != ErrKeyNotExists {
		return
	}

	log.I.Ln("key not provided, generating a new one.")

	tunKey, _ = NewPrivateKey()

	o.store.SetKey(tunKey)

	log.I.Ln("rpc tunnel public key:")
	log.I.Ln("-", tunKey.PubKey().Encode())
}

func configureTunnelPort() {

	if o.tunPort != NullPort {
		return
	}

	log.I.Ln("rpc tunnel port not provided, generating a random one.")

	o.tunPort = genRandomPort(10000)
}

func configurePeerWhitelist() {

	if len(o.tunPeers) == 0 {
		return
	}

	log.I.Ln("rpc tunnel whitelisted peers:")

	for _, peer := range o.tunPeers {

		var pubKey RPCPublicKey

		pubKey.Decode(peer)

		log.I.Ln("-", pubKey.Encode())

		tunWhitelist = append(tunWhitelist, pubKey)
	}
}
