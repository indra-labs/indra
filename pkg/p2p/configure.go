package p2p

import (
	"github.com/multiformats/go-multiaddr"
	"github.com/spf13/viper"
)

func configure() {
	configureSeeds()
}

func configureKey() {

	if viper.GetString(keyFlag) == "" {

	}

}

func configureListeners() {

	if len(viper.GetString(listenFlag)) > 0 {

	}

}

func configureSeeds() {

	if len(viper.GetStringSlice(connectFlag)) > 0 {

		log.I.Ln("connect only detected, using only the connect seed addresses")

		for _, connector := range viper.GetStringSlice(connectFlag) {
			seedAddresses = append(seedAddresses, multiaddr.StringCast(connector))
		}

		return
	}

	var err error

	if seedAddresses, err = netParams.ParseSeedMultiAddresses(); err != nil {
		return
	}

	if len(viper.GetStringSlice("seed")) > 0 {

		log.I.Ln("found", len(viper.GetStringSlice("seed")), "additional seeds.")

		for _, seed := range viper.GetStringSlice("seed") {
			seedAddresses = append(seedAddresses, multiaddr.StringCast(seed))
		}
	}

	return
}
