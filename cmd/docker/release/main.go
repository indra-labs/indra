package main

import (
	"context"
	"github.com/Indra-Labs/indra"
	"github.com/Indra-Labs/indra/pkg/docker"
	"github.com/cybriq/proc/pkg/app"
	"github.com/cybriq/proc/pkg/cmds"
	log2 "github.com/cybriq/proc/pkg/log"
	"github.com/cybriq/proc/pkg/opts/config"
	"github.com/cybriq/proc/pkg/opts/meta"
	"github.com/cybriq/proc/pkg/opts/toggle"
	"github.com/cybriq/proc/pkg/path"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"os"
	"time"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

func init() {
	log2.App = "indra"
}

var (
	timeout = 120 * time.Second
)

var commands = &cmds.Command{
	Name:          "release",
	Description:   "Builds the indra docker image and pushes it to a list of docker repositories.",
	Documentation: lorem,
	Default: cmds.Tags("release"),
	Configs:       config.Opts{
		"stable": toggle.New(meta.Data{
			Label:         "stable",
			Description:   "tag the current build as stable.",
			Documentation: lorem,
			Default: "false",
		}),
		"push": toggle.New(meta.Data{
			Label:         "push",
			Description:   "push the newly built/tagged images to the docker repositories.",
			Documentation: lorem,
			Default: "false",
		}),
	},
	Entrypoint: func(command *cmds.Command, args []string) error {

		var err error
		var cli *client.Client

		if cli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation()); check(err) {
			return err
		}

		defer cli.Close()

		// Set a Timeout for 120 seconds
		ctx, cancel := context.WithTimeout(context.Background(), timeout)

		defer cancel()

		// If we've flagged stable, we should also build a stable tag
		if command.GetOpt(path.From("release stable")).Value().Bool() {
			docker.SetRelease()
		}

		var builder = docker.NewBuilder(ctx, cli)

		if err = builder.Build(); check(err) {
			return err
		}

		if command.GetOpt(path.From("release push")).Value().Bool() {
			if err = builder.Push(types.ImagePushOptions{}); check(err) {
				return err
			}
		}

		if err = builder.Close(); check(err) {
			return err
		}

		return nil
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