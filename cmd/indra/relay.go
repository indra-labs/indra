package main

//
// import (
//	"github.com/spf13/cobra"
//	"github.com/spf13/viper"
//
//	"github.com/indra-labs/indra"
//	"github.com/indra-labs/indra/pkg/crypto/key/prv"
//	"github.com/indra-labs/indra/pkg/crypto/key/pub"
//	"github.com/indra-labs/indra/pkg/interrupt"
//	log2 "github.com/indra-labs/indra/pkg/proc/log"
//	"github.com/indra-labs/indra/pkg/relay"
//	"github.com/indra-labs/indra/pkg/relay/transport"
//	"github.com/indra-labs/indra/pkg/util/slice"
// )
//
// var (
//	eng       *relay.Engine
//	engineP2P []string
//	engineRPC []string
// )
//
// func init() {
//	pf := relayCmd.PersistentFlags()
//	pf.StringSliceVarP(&engineP2P, "engineP2P-relay", "P",
//		[]string{"127.0.0.1:8337", "::1:8337"},
//		"address/ports for IPv4 and v6 listeners")
//	pf.StringSliceVarP(&engineRPC, "relay-control", "r",
//		[]string{"127.0.0.1:8339", "::1:8339"},
//		"address/ports for IPv4 and v6 listeners")
//	viper.BindPFlag("engineP2P-relay", seedCommand.PersistentFlags().Lookup("engineP2P-relay"))
//	viper.BindPFlag("relay-control", seedCommand.PersistentFlags().Lookup(
//		"relay-control"))
//	rootCmd.AddCommand(relayCmd)
// }
//
// var relayCmd = &cobra.Command{
//	Use:   "relay",
//	Short: "Runs a relay server.",
//	Long:  `Runs a server that can be controlled with RPC and CLI interfaces.`,
//	Run: func(cmd *cobra.Command, args []string) {
//
//		log2.SetLogLevel(log2.Debug)
//
//		log.I.Ln("-- ", log2.App, "("+viper.GetString(""+
//			"network")+") -",
//			indra.SemVer, "- Network Freedom. --")
//
//		var e error
//		var idPrv *prv.HiddenService
//		if idPrv, e = prv.GenerateKey(); check(e) {
//			return
//		}
//		idPub := pub.Derive(idPrv)
//		nTotal := 5
//		tpt := transport.NewSim(nTotal)
//		addr := slice.GenerateRandomAddrPortIPv4()
//		nod, _ := relay.NewNode(addr, idPub, idPrv, tpt, 50000, true)
//		eng, e = relay.NewEngine(tpt, idPrv, nod, nil, 5)
//		interrupt.AddHandler(func() { eng.Q() })
//		log.D.Ln("starting up server")
//		eng.Start()
//		eng.Wait()
//		log.I.Ln("fin")
//		return
//	},
// }
