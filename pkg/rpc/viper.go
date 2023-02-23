package rpc

import (
	"github.com/multiformats/go-multiaddr"
	"github.com/spf13/viper"
	"strconv"
)

func ConfigureWithViper() (err error) {

	log.I.Ln("initializing the rpc server")

	config.SetKey(viper.GetString("rpc-key"))

	if config.IsNullKey() {

		log.I.Ln("rpc key not provided, generating a new one.")

		config.NewKey()
	}

	log.I.Ln("rpc public key:")
	log.I.Ln("-", config.Key.PubKey().Encode())

	config.SetUnixPath(viper.GetString("rpc-listen-unix"))
	config.SetPort(viper.GetUint16("rpc-listen-port"))

	if viper.GetUint16("rpc-listen-port") == NullPort {

		viper.Set("rpc-listen-port", config.SetRandomPort())
	}

	log.I.Ln("rpc listeners:")
	log.I.Ln("- [/ip4/0.0.0.0/udp/"+strconv.Itoa(int(config.ListenPort)), "/ip6/:::/udp/"+strconv.Itoa(int(config.ListenPort))+" /unix"+config.UnixPath+"]")

	for _, ip := range viper.GetStringSlice("rpc-whitelist-deviceIP") {
		config.IP_Whitelist = append(config.IP_Whitelist, multiaddr.StringCast(ip))
	}

	for _, peer := range viper.GetStringSlice("rpc-whitelist-peer") {

		var pubKey RPCPublicKey

		pubKey.Decode(peer)

		config.Peer_Whitelist = append(config.Peer_Whitelist, pubKey)
	}

	return
}
