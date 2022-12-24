package main

import (
	"fmt"
	"github.com/Indra-Labs/indra"
	"github.com/Indra-Labs/indra/pkg/server"
	"github.com/cybriq/proc/pkg/cmds"
)

var defaultHandler = func(c *cmds.Command, args []string) error {
	fmt.Println("indra.")
	return nil
}

var versionHandler = func(c *cmds.Command, args []string) error {
	fmt.Println(indra.SemVer)
	return nil
}

var serveHandler = func(c *cmds.Command, args []string) error {

	log.I.Ln("running serve.")

	var err error
	var srv *server.Server

	if srv, err = server.New(server.DefaultServerConfig); check(err) {
		return err
	}

	log.I.Ln("starting the server.")

	if srv.Serve(); check(err) {
		return err
	}

	return nil
}

var cliHandler = func(c *cmds.Command, args []string) error {
	fmt.Println(indra.SemVer)
	return nil
}
