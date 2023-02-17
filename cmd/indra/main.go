package main

import (
	"git-indra.lan/indra-labs/indra"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"os"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

func init() {
	log2.App = "indra"
}

func main() {

	var err error

	if err = rootCmd.Execute(); check(err) {
		os.Exit(1)
	}

	os.Exit(0)
}
