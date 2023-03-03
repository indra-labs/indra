package rpc

import (
	"github.com/spf13/viper"
)

func configureUnixSocket() {

	unixPath = viper.GetString(unixPathFlag)

	if unixPath == "" {
		return
	}

	log.I.Ln("enabling rpc unix listener:")
	log.I.F("- [/unix%s]", unixPath)

	isUnixSockEnabled = true
}

func configureTunnel() {

	isTunnelEnabled = viper.GetBool(tunEnableFlag)

	if !isTunnelEnabled {
		return
	}

	log.I.Ln("enabling rpc tunnel")

	configureTunnelPort()

	log.I.Ln("rpc tunnel listeners:")
	log.I.F("- [/ip4/0.0.0.0/udp/%d /ip6/:::/udp/%d]", viper.GetUint16(tunPortFlag), viper.GetUint16(tunPortFlag))

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

	if viper.GetUint16(tunPortFlag) != NullPort {

		tunnelPort = int(viper.GetUint16(tunPortFlag))

		return
	}

	log.I.Ln("rpc tunnel port not provided, generating a random one.")

	viper.Set(tunPortFlag, genRandomPort(10000))

	tunnelPort = int(viper.GetUint16(tunPortFlag))
}

func configurePeerWhitelist() {

	if len(viper.GetStringSlice(tunPeersFlag)) == 0 {
		return
	}

	log.I.Ln("rpc tunnel whitelisted peers:")

	for _, peer := range viper.GetStringSlice(tunPeersFlag) {

		var pubKey RPCPublicKey

		pubKey.Decode(peer)

		log.I.Ln("-", pubKey.Encode())

		tunWhitelist = append(tunWhitelist, pubKey)
	}
}
