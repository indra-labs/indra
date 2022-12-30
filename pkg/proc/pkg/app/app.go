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
	// Add the default configuration items for datadir/configfile
	cmds.GetConfigBase(cmd.Configs, cmd.Name, false)
	// Add the help function
	cmd.AddCommand(cmds.Help())
	a = &App{Command: cmd}
	// We first parse the CLI args, in case config file location has been
	// specified
	if a.launch, _, err = a.Command.ParseCLIArgs(args); log.E.Chk(err) {
		return
	}
	if err = cmd.LoadConfig(); log.E.Chk(err) {
		return
	}
	a.Command, err = cmds.Init(cmd, nil)
	a.Envs = cmd.GetEnvs()
	if err = a.Envs.LoadFromEnvironment(); log.E.Chk(err) {
		return
	}
	// This is done again, to ensure the effect of CLI args take precedence
	if a.launch, a.runArgs, err = a.Command.ParseCLIArgs(args); log.E.Chk(err) {
		return
	}
	return
}

func (a *App) Launch() (err error) {
	err = a.launch.Entrypoint(a.Command, a.runArgs)
	log.E.Chk(err)
	return
}
