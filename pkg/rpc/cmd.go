package rpc

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	flag_tun_enable        = "rpc-tun-enable"
	flag_tun_key           = "rpc-tun-key"
	flag_tun_port          = "rpc-tun-port"
	flag_tun_whitlist_peer = "rpc-tun-whitelist-peer"
)

var (
	rpc_tun_enable         bool
	rpc_tun_key            string
	rpc_tun_port           uint16
	rpc_tun_whitelist_peer []string
)

func Configure(cmd *cobra.Command) {

	defineUnixSocket(cmd)

	cmd.PersistentFlags().BoolVarP(&rpc_tun_enable, flag_tun_enable, "", false, "enables the rpc server tunnel")
	cmd.PersistentFlags().Uint16VarP(&rpc_tun_port, flag_tun_port, "", 0, "binds the udp server to port (random if not selected)")
	cmd.PersistentFlags().StringVarP(&rpc_tun_key, flag_tun_key, "", "", "the base58 encoded pre-shared key for accessing the rpc")
	cmd.PersistentFlags().StringSliceVarP(&rpc_tun_whitelist_peer, flag_tun_whitlist_peer, "", []string{}, "adds a peer id to the whitelist for access")

	viper.BindPFlag(flag_tun_enable, cmd.PersistentFlags().Lookup(flag_tun_enable))
	viper.BindPFlag(flag_tun_port, cmd.PersistentFlags().Lookup(flag_tun_port))
	viper.BindPFlag(flag_tun_key, cmd.PersistentFlags().Lookup(flag_tun_key))
	viper.BindPFlag(flag_tun_whitlist_peer, cmd.PersistentFlags().Lookup(flag_tun_whitlist_peer))
}

func ConfigureWithViper() (err error) {

	log.I.Ln("configuring the rpc server")

	configureUnixSocket()
	configureTunnel()

	log.I.Ln("rpc listeners:")
	log.I.F("- [/ip4/0.0.0.0/udp/%d", config.listenPort)
	log.I.F("/ip4/0.0.0.0/udp/%d", config.listenPort)
	log.I.F("/ip6/:::/udp/%d", config.listenPort)
	log.I.F("/unix" + unixPath + "]")

	return
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

var (
	tunKey *RPCPrivateKey
)

func configureTunnelKey() {

	if viper.GetString("rpc-tun-key") == "" {

		log.I.Ln("rpc tunnel key not provided, generating a new one.")

		tunKey, _ = NewPrivateKey()

		viper.Set("rpc-tun-key", tunKey.Encode())
	}

	log.I.Ln("rpc public key:")
	log.I.Ln("-", config.key.PubKey().Encode())
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

		config.peerWhitelist = append(config.peerWhitelist, pubKey)
	}
}
