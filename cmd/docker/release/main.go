// Package release is a tool to create and publish docker images.
//
// Currently only has the LND build implemented.
package main

import (
	"context"

	"github.com/docker/docker/client"

	"git.indra-labs.org/dev/ind/pkg/docker"
	"git.indra-labs.org/dev/ind/pkg/proc/app"
	"git.indra-labs.org/dev/ind/pkg/proc/cmds"
	log2 "git.indra-labs.org/dev/ind/pkg/proc/log"
	"git.indra-labs.org/dev/ind/pkg/proc/opts/config"
	"git.indra-labs.org/dev/ind/pkg/proc/opts/meta"
	"git.indra-labs.org/dev/ind/pkg/proc/opts/toggle"

	"os"
	"time"
)

var (
	log   = log2.GetLogger()
	check = log.E.Chk
)

func init() {
	log2.App.Store("indra")
}

var (
	defaultBuildingTimeout = 800 * time.Second
	defaultRepositoryName  = "indralabs"
)

func strPtr(str string) *string { return &str }

var commands = &cmds.Command{
	Name:          "release",
	Description:   "Builds the indra docker image and pushes it to a list of docker repositories.",
	Documentation: lorem,
	Default:       cmds.Tags("release"),
	Configs: config.Opts{
		"stable": toggle.New(meta.Data{
			Label:         "stable",
			Description:   "tag the current build as stable.",
			Documentation: lorem,
			Default:       "false",
		}),
		"push": toggle.New(meta.Data{
			Label:         "push",
			Description:   "push the newly built/tagged images to the docker repositories.",
			Documentation: lorem,
			Default:       "false",
		}),
	},
	Entrypoint: func(command *cmds.Command, args []string) error {

		// If we've flagged stable, we should also build a stable tag
		if command.GetValue("stable").Bool() {
			docker.SetRelease()
		}

		// If we've flagged push, the tags will be pushed to all repositories.
		if command.GetValue("push").Bool() {
			docker.SetPush()
		}

		// Set a Timeout for 120 seconds
		ctx, cancel := context.WithTimeout(context.Background(), defaultBuildingTimeout)
		defer cancel()

		// Setup a new instance of the docker client

		var err error
		var cli *client.Client

		if cli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation()); check(err) {
			return err
		}

		defer cli.Close()

		// Get ready to submit the builds
		var builder = docker.NewBuilder(ctx, cli, sourceConfigurations, buildConfigurations, packagingConfigurations)

		if err = builder.Build(); check(err) {
			return err
		}

		if err = builder.Push(); check(err) {
			return err
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
