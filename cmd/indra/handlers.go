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

	log.I.Ln("-- “Far away in the heavenly abode of the great god indra, there is a wonderful net which has been hung by some cunning artificer in such a manner that it stretches out indefinitely in all directions.”")
	log.I.Ln("-- “In accordance with the extravagant tastes of deities, the artificer has hung a single glittering jewel at the net’s every node, and since the net itself is infinite in dimension, the jewels are infinite in number. There hang the jewels, glittering like stars of the first magnitude, a wonderful sight to behold.”")
	log.I.Ln("-- “If we now arbitrarily select one of these jewels for inspection and look closely at it, we will discover that in its polished surface there are reflected all the other jewels in the net, infinite in number. Not only that, but each of the jewels reflected in this one jewel is also reflecting all the other jewels, so that the process of reflection is infinite.”")

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
