package main

import (
	"github.com/Indra-Labs/indra"
	"github.com/cybriq/proc/pkg/app"
	log2 "github.com/cybriq/proc/pkg/log"
	"os"
)

var (
	log      = log2.GetLogger(indra.PathBase)
	check    = log.E.Chk
)

func main() {

	log2.App = "indra"

	log.I.Ln("indra - the indra network tool")
	log.I.Ln("version:", indra.SemVer)

	var err error
	var application *app.App

	if application, err = app.New(commands, os.Args); check(err) {
		os.Exit(1)
	}

	if err = application.Launch(); check(err) {
		os.Exit(1)
	}

	log.D.Ln("Shutdown complete.")
}
