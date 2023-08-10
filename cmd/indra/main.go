// Indra is a low latency, source routed mixnet distributed virtual private network protocol.
//
// This application includes a client, client GUI, relay and seed node according to the subcommand invocation.
package main

import (
	"os"

	log2 "git.indra-labs.org/dev/ind/pkg/proc/log"
)

var (
	log   = log2.GetLogger()
	check = log.E.Chk
)

func init() {
	log2.App.Store("indra")
}

func main() {

	var err error

	if err = rootCmd.Execute(); check(err) {
		os.Exit(1)
	}

	os.Exit(0)
}
