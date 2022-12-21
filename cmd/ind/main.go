package main

import (
	"github.com/cybriq/proc/pkg/app"
	"github.com/cybriq/proc/pkg/cmds"
	"github.com/cybriq/proc/pkg/opts/config"
	"github.com/cybriq/proc/pkg/opts/meta"
	"github.com/cybriq/proc/pkg/opts/toggle"
	"github.com/davecgh/go-spew/spew"
	"os"
)

const lorem = `
Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor
incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis 
nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. 
Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu 
fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in 
culpa qui officia deserunt mollit anim id est laborum.`

func main() {

	command := &cmds.Command{
		Name:          "ind",
		Description:   "The indra network daemon.",
		Documentation: lorem,
		Configs: config.Opts{
			"AutoPorts": toggle.New(meta.Data{
				Label:         "Automatic Ports",
				Tags:          cmds.Tags("node", "wallet"),
				Description:   "RPC and controller ports are randomized, use with controller for automatic peer discovery",
				Documentation: lorem,
				Default:       "false",
			}),
		},
		Commands: cmds.Commands{
			{
				Name:          "version",
				Description:   "print version and exit",
				Documentation: lorem,
			},
		},
	}

	cmds.GetConfigBase(command.Configs, command.Name, false)

	cmds.Init(command, os.Args)

	var application *app.App
	var err error

	spew.Dump(command)

	if application, err = app.New(command, os.Args); err != nil {
		spew.Dump(err)
		os.Exit(1)
	}

	spew.Dump(application)

	if err = application.Launch(); err != nil {
		spew.Dump(err)
		os.Exit(1)
	}
}
