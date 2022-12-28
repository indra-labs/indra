package main

import (
	"github.com/cybriq/proc/pkg/cmds"
	"github.com/cybriq/proc/pkg/opts/config"
)

var (
	commands = &cmds.Command{
		Name:          "indra",
		Description:   "Network Freedom.",
		Documentation: lorem,
		Entrypoint:    defaultHandler,
		Default:       cmds.Tags("help"),
		Configs: config.Opts{
			//"AutoPorts": toggle.New(meta.Data{
			//	Label:         "Automatic Ports",
			//	Tags:          cmds.Tags("node", "wallet"),
			//	Description:   "RPC and controller ports are randomized, use with controller for automatic peer discovery",
			//	Documentation: lorem,
			//	Default:       "false",
			//}),
		},
		Commands: cmds.Commands{
			{
				Name:        "version",
				Description: "print indra version",

				Documentation: lorem,
				Entrypoint: versionHandler,
			},
			{
				Name:        "cli",
				Description: "a command line client for managing an indra network daemon",

				Documentation: lorem,
				Entrypoint: cliHandler,
			},
			{
				Name:        "serve",
				Description: "serves an instance of the indra network daemon",

				Documentation: lorem,
				Entrypoint: serveHandler,
			},
		},
	}
)

const lorem = `
Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor
incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis 
nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. 
Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu 
fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in 
culpa qui officia deserunt mollit anim id est laborum.`
