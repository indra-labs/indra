package main

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	
	"git-indra.lan/indra-labs/indra"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/interrupt"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/relay"
	"git-indra.lan/indra-labs/indra/pkg/transport"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
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
		
		log2.SetLogLevel(log2.Trace)
		
		log.I.Ln("-- ", log2.App, "("+viper.GetString(""+
			"network")+") -",
			indra.SemVer, "- Network Freedom. --")
		
		var e error
		var idPrv *prv.Key
		if idPrv, e = prv.GenerateKey(); check(e) {
			return
		}
		idPub := pub.Derive(idPrv)
		nTotal := 5
		tpt := transport.NewSim(nTotal)
		addr := slice.GenerateRandomAddrPortIPv4()
		nod, _ := relay.NewNode(addr, idPub, idPrv, tpt, 50000, true)
		eng, e = relay.NewEngine(tpt, idPrv, nod, nil, 5)
		interrupt.AddHandler(func() { eng.Q() })
		log.D.Ln("starting up server")
		eng.Start()
		eng.Wait()
		log.I.Ln("fin")
		return
	},
}
