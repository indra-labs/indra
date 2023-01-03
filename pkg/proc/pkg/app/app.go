package app

import (
	"github.com/cybriq/proc/pkg/cmds"
)

type App struct {
	*cmds.Command
	launch  *cmds.Command
	runArgs []string
	cmds.Envs
}

func New(cmd *cmds.Command, args []string) (a *App, err error) {

	var app App

	app.Command = cmd

	// Add the default configuration items for datadir/configfile
	cmds.GetConfigBase(cmd.Configs, cmd.Name, false)

	// Add the help function
	cmd.AddCommand(cmds.Help())

	// We first parse the CLI args, in case config file location has been
	// specified
	if app.launch, _, err = app.Command.ParseCLIArgs(args); log.E.Chk(err) {
		return
	}

	if err = cmd.LoadConfig(); log.E.Chk(err) {
		return
	}

	app.Command, err = cmds.Init(cmd, nil)

	// Load the environment variables
	app.Envs = cmd.GetEnvs()
	if err = app.Envs.LoadFromEnvironment(); log.E.Chk(err) {
		return
	}

	// This is done again, to ensure the effect of CLI args take precedence
	if app.launch, app.runArgs, err = app.Command.ParseCLIArgs(args); log.E.Chk(err) {
		return
	}

	return &app, nil
}

func (a *App) Launch() (err error) {
	err = a.launch.Entrypoint(a.launch, a.runArgs)
	log.E.Chk(err)
	return
}
