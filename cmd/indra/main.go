package main

import (
	"os"
	
	"github.com/indra-labs/indra"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
)

var (
	log   = log2.GetLogger(indra.PathBase)
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
