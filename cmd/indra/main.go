package main

import (
	"fmt"
	"os"

	"github.com/indra-labs/indra"
	"github.com/indra-labs/indra/pkg/app"
	"github.com/indra-labs/indra/pkg/cfg"
	"github.com/indra-labs/indra/pkg/cmds"
	log2 "github.com/indra-labs/indra/pkg/log"
	"github.com/indra-labs/indra/pkg/opts/config"
	"github.com/indra-labs/indra/pkg/opts/list"
	"github.com/indra-labs/indra/pkg/opts/meta"
	"github.com/indra-labs/indra/pkg/opts/text"
	"github.com/indra-labs/indra/pkg/server"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/multiformats/go-multiaddr"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

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
				"key": text.New(meta.Data{
					Label:         "key",
					Description:   "A base58 encoded private key.",
					Documentation: lorem,
				}),
				"seed": list.New(meta.Data{
					Label:         "seed",
					Description:   "Adds additional seeds by multiaddress. Examples: /dns4/seed0.example.com/tcp/8337, /ip4/127.0.0.1/tcp/8337",
					Documentation: lorem,
				}, multiAddrSanitizer),
				"peer": list.New(meta.Data{
					Label:         "peer",
					Description:   "Adds a list of peer multiaddresses. Example: /ip4/0.0.0.0/tcp/8337",
					Documentation: lorem,
				}, multiAddrSanitizer),
				"listen": list.New(meta.Data{
					Label:         "listen",
					Description:   "A list of listener multiaddresses. Example: /ip4/0.0.0.0/tcp/8337",
					Documentation: lorem,
					Default:       "/ip4/127.0.0.1/tcp/8337,/ip6/::1/tcp/8337",
				}, multiAddrSanitizer),
			},
			Entrypoint: func(c *cmds.Command, args []string) error {

				var err error
				var params = cfg.SimnetServerParams

				log.I.Ln("-- ", log2.App, "("+params.Name+") -", indra.SemVer, "- Network Freedom. --")

				var privKey crypto.PrivKey

				if privKey, err = server.Base58Decode(c.GetValue("key").Text()); check(err) {
					return err
				}

				server.DefaultConfig.PrivKey = privKey

				for _, listener := range c.GetListValue("listen") {
					server.DefaultConfig.ListenAddresses = append(server.DefaultConfig.ListenAddresses, multiaddr.StringCast(listener))
				}

				for _, seed := range c.GetListValue("seed") {
					server.DefaultConfig.SeedAddresses = append(server.DefaultConfig.SeedAddresses, multiaddr.StringCast(seed))
				}

				var srv *server.Server

				log.I.Ln("running serve.")

				if srv, err = server.New(params, server.DefaultConfig); check(err) {
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

func multiAddrSanitizer(opt *list.Opt) error {

	// log.I.Ln("adding p2p listener", opt.String())

	return nil
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
