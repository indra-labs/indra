package main

import (
	"fmt"
	"os"

	"github.com/Indra-Labs/indra"
	"github.com/Indra-Labs/indra/pkg/app"
	"github.com/Indra-Labs/indra/pkg/cmds"
	log2 "github.com/Indra-Labs/indra/pkg/log"
	"github.com/Indra-Labs/indra/pkg/opts/config"
	"github.com/Indra-Labs/indra/pkg/opts/list"
	"github.com/Indra-Labs/indra/pkg/opts/meta"
	"github.com/Indra-Labs/indra/pkg/server"
	"github.com/davecgh/go-spew/spew"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

func init() {
	log2.App = "indra"
}

var commands = &cmds.Command{
	Name:          "indra",
	Description:   "Network Freedom.",
	Documentation: lorem,
	Default:       cmds.Tags("help"),
	Configs:       config.Opts{},
	Entrypoint: func(c *cmds.Command, args []string) error {

		fmt.Println("indra")

		return nil
	},
	Commands: cmds.Commands{
		{
			Name:          "version",
			Description:   "prints the indra version",
			Documentation: lorem,
			Configs:       config.Opts{},
			Entrypoint: func(c *cmds.Command, args []string) error {

				fmt.Println(indra.SemVer)

				return nil
			},
		},
		{
			Name:          "cli",
			Description:   "a command line client for managing an indra network daemon",
			Documentation: lorem,
			Configs:       config.Opts{},
			Entrypoint: func(c *cmds.Command, args []string) error {

				fmt.Println(indra.SemVer)

				return nil
			},
		},
		{
			Name:          "serve",
			Description:   "serves an instance of the indra network daemon",
			Documentation: lorem,
			Configs: config.Opts{
				"seed": list.New(meta.Data{
					Label:         "seed",
					Description:   "Adds additional seeds by hostname, or multiaddress. Examples: seed0.example.com, /ip4/127.0.0.1/tcp/8337",
					Documentation: lorem,
					Default:       "/ip4/172.16.238.2/tcp/8337",
				}, func(opt *list.Opt) error {

					log.I.Ln("adding seed", opt.String())

					return nil
				}),
				"peer": list.New(meta.Data{
					Label:         "peer",
					Description:   "Adds a list of peer multiaddresses. Example: /ip4/0.0.0.0/tcp/8337",
					Documentation: lorem,
					Default:       "/ip4/0.0.0.0/tcp/8337",
				}, func(opt *list.Opt) error {

					log.I.Ln("adding peer", opt.String())

					return nil
				}),
				"listen": list.New(meta.Data{
					Label:         "listen",
					Description:   "A list of listener multiaddresses. Example: /ip4/0.0.0.0/tcp/8337",
					Documentation: lorem,
					Default:       "/ip4/0.0.0.0/tcp/8337",
				}, func(opt *list.Opt) error {

					log.I.Ln("adding p2p listener", opt.String())

					return nil
				}),
			},
			Entrypoint: func(c *cmds.Command, args []string) error {

				log.I.Ln("-- ", log2.App, "-", indra.SemVer, "- Network Freedom. --")

				spew.Dump(c.Configs)

				var err error
				var srv *server.Server

				log.I.Ln("running serve.")

				if srv, err = server.New(server.DefaultServerConfig); check(err) {
					return err
				}

				log.I.Ln("starting the server.")

				if srv.Serve(); check(err) {
					return err
				}

				log.I.Ln("-- fin --")

				return nil
			},
		},
	},
}

func main() {

	var err error
	var application *app.App

	// Creates a new application
	if application, err = app.New(commands, os.Args); check(err) {
		os.Exit(1)
	}

	// Launches the application
	if err = application.Launch(); check(err) {
		os.Exit(1)
	}

	os.Exit(0)
}

const lorem = `
Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor
incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis 
nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. 
Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu 
fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in 
culpa qui officia deserunt mollit anim id est laborum.`
