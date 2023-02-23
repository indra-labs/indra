package rpc

import (
	"github.com/spf13/viper"
)

func ConfigureWithViper() (err error) {

	log.I.Ln("configuring the rpc server")

	configureUnixSocket()
	configureTunnel()

	log.I.Ln("rpc listeners:")
	log.I.F("- [/ip4/0.0.0.0/udp/%d", devicePort)
	log.I.F("/ip4/0.0.0.0/udp/%d", devicePort)
	log.I.F("/ip6/:::/udp/%d", devicePort)
	log.I.F("/unix" + unixPath + "]")

	return
}

func configureUnixSocket() {

	if viper.GetString(unixPathFlag) == "" {
		return
	}

	log.I.Ln("enabling unix listener:", viper.GetString(unixPath))

	isUnixSockEnabled = true
}

func configureTunnel() {

	if !viper.GetBool("rpc-tun-enable") {
		return
	}

	log.I.Ln("enabling rpc tunnel")

	configureTunnelKey()
	configureTunnelPort()
	configurePeerWhitelist()

	enableTunnel()
}

func configureTunnelKey() {

	if viper.GetString("rpc-tun-key") == "" {

		log.I.Ln("rpc tunnel key not provided, generating a new one.")

		tunKey, _ = NewPrivateKey()

		viper.Set("rpc-tun-key", tunKey.Encode())
	}

	log.I.Ln("rpc public key:")
	log.I.Ln("-", tunKey.PubKey().Encode())
}

func configureTunnelPort() {

	if viper.GetUint16("rpc-tun-port") == NullPort {

		log.I.Ln("rpc tunnel port not provided, generating a random one.")

		viper.Set("rpc-tun-port", genRandomPort(10000))
	}
}

func configurePeerWhitelist() {
	for _, peer := range viper.GetStringSlice("rpc-whitelist-peer") {

		var pubKey RPCPublicKey

		pubKey.Decode(peer)

		tunWhitelist = append(tunWhitelist, pubKey)
	}
}
