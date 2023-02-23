package rpc

import (
	"context"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"os"
)

func RunWith(ctx context.Context, r func(srv *grpc.Server)) {

	log.I.Ln("initializing the rpc server")

	var err error

	if err = configureWithViper(); check(err) {
		os.Exit(1)
	}

	r(server)

	log.I.Ln("starting rpc server")

	go Start(ctx)
}

func configureWithViper() (err error) {

	log.I.Ln("configuring the rpc server")

	configureUnixSocket()
	configureTunnel()

	return
}

func configureUnixSocket() {

	if viper.GetString(unixPathFlag) == "" {
		return
	}

	log.I.F("enabling rpc unix listener  [/unix%s]", viper.GetString(unixPathFlag))

	isUnixSockEnabled = true
}

func configureTunnel() {

	if !viper.GetBool(tunEnableFlag) {

		log.I.Ln("disabling rpc tunnel")

		return
	}

	enableTunnel()

	log.I.Ln("enabling rpc tunnel")

	configureTunnelPort()

	log.I.Ln("rpc tunnel listeners:")
	log.I.F("- [/ip4/0.0.0.0/udp/%d /ip6/:::/udp/%d]", viper.GetUint16(tunPortFlag), viper.GetUint16(tunPortFlag))

	configureTunnelKey()

	configurePeerWhitelist()

}

func configureTunnelKey() {

	if viper.GetString(tunKeyFlag) == "" {

		log.I.Ln("rpc tunnel key not provided, generating a new one.")

		tunKey, _ = NewPrivateKey()

		viper.Set(tunKeyFlag, tunKey.Encode())
	}

	tunKey = &RPCPrivateKey{}
	tunKey.Decode(viper.GetString(tunKeyFlag))

	log.I.Ln("rpc tunnel public key:")
	log.I.Ln("-", tunKey.PubKey().Encode())
}

func configureTunnelPort() {

	if viper.GetUint16(tunPortFlag) != NullPort {
		return
	}

	log.I.Ln("rpc tunnel port not provided, generating a random one.")

	viper.Set(tunPortFlag, genRandomPort(10000))
}

func configurePeerWhitelist() {

	if len(viper.GetStringSlice(tunPeersFlag)) == 0 {
		return
	}

	log.I.Ln("rpc tunnel whitelisted peers:")

	for _, peer := range viper.GetStringSlice(tunPeersFlag) {

		var pubKey RPCPublicKey

		pubKey.Decode(peer)

		tunWhitelist = append(tunWhitelist, pubKey)

		log.I.Ln("-", pubKey.Encode())

	}
}
