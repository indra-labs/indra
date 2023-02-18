package main

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	
	"git-indra.lan/indra-labs/indra"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/relay"
)

var (
	eng *relay.Engine
	p2p []string
	rpc []string
)

func init() {
	pf := relayCmd.PersistentFlags()
	pf.StringSliceVarP(&p2p, "p2p-relay", "P",
		[]string{"127.0.0.1:8337", "::1:8337"},
		"address/ports for IPv4 and v6 listeners")
	pf.StringSliceVarP(&rpc, "relay-control", "r",
		[]string{"127.0.0.1:8339", "::1:8339"},
		"address/ports for IPv4 and v6 listeners")
	viper.BindPFlag("p2p-relay", seedCmd.PersistentFlags().Lookup("p2p-relay"))
	viper.BindPFlag("relay-control", seedCmd.PersistentFlags().Lookup(
		"relay-control"))
	rootCmd.AddCommand(relayCmd)
}

var relayCmd = &cobra.Command{
	Use:   "relay",
	Short: "Runs a relay server.",
	Long:  `Runs a server that can be controlled with RPC and CLI interfaces.`,
	Run: func(cmd *cobra.Command, args []string) {
		
		log.I.Ln("-- ", log2.App, "("+viper.GetString("network")+") -",
			indra.SemVer, "- Network Freedom. --")
		
		log.I.Ln("fin")
		return
	},
}
