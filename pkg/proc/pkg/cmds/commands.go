package cmds

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	log2 "github.com/cybriq/proc/pkg/log"
	"github.com/cybriq/proc/pkg/opts/config"
	"github.com/cybriq/proc/pkg/opts/meta"
	"github.com/cybriq/proc/pkg/opts/text"
	"github.com/cybriq/proc/pkg/opts/toggle"
	"github.com/cybriq/proc/pkg/path"
	"github.com/cybriq/proc/pkg/util"
)

type Op func(c *Command, args []string) error

var NoOp = func(c *Command, args []string) error { return nil }
var Tags = func(s ...string) []string {
	return s
}

// Command is a specification for a command and can include any number of
// subcommands, and for each Command a list of options
type Command struct {
	path.Path
	Name          string
	Description   string
	Documentation string
	Entrypoint    Op
	Parent        *Command
	Commands      Commands
	Configs       config.Opts
	Default       []string // specifies default subcommand to execute
	sync.Mutex
}

// Commands are a slice of Command entries
type Commands []*Command

func (c *Command) AddCommand(cm *Command) {
	c.Commands = append(c.Commands, cm)
}

const configFilename = "config.toml"

// GetConfigBase creates an option set that should go in the root of a
// Command specification for an application, providing a data directory path
// and config file path.
//
// This exists in order to simplify setup for application configuration
// persistence.
func GetConfigBase(in config.Opts, appName string, abs bool) {
	var defaultDataDir, defaultConfigFile string
	switch runtime.GOOS {
	case "linux", "aix", "freebsd", "netbsd", "openbsd", "dragonfly":
		defaultDataDir = fmt.Sprintf("~/.%s", appName)
		defaultConfigFile =
			fmt.Sprintf("%s/%s", defaultDataDir, configFilename)
	case "windows":
		defaultDataDir = fmt.Sprintf("%%LOCALAPPDATA%%\\%s", appName)
		defaultConfigFile =
			fmt.Sprintf("%%LOCALAPPDATA%%\\%s\\%s", defaultDataDir,
				configFilename)
	case "darwin":
		defaultDataDir = filepath.Join(
			"~", "Library",
			"Application Support", strings.ToUpper(appName),
		)
		defaultConfigFile = filepath.Join(defaultDataDir, configFilename)
	}
	options := config.Opts{
		"ConfigFile": text.New(meta.Data{
			Aliases:     []string{"CF"},
			Label:       "Configuration File",
			Description: "location of configuration file",
			Documentation: strings.TrimSpace(`
The configuration file path defines the place where the configuration will be
loaded from at application startup, and where it will be written if changed.
`),
			Default: defaultConfigFile,
		}, text.NormalizeFilesystemPath(abs, appName)),

		"DataDir": text.New(meta.Data{
			Aliases:     []string{"DD"},
			Label:       "Data Directory",
			Description: "root folder where application data is stored",
			Default:     defaultDataDir,
		}, text.NormalizeFilesystemPath(abs, appName)),

		"LogCodeLocations": toggle.New(meta.Data{
			Aliases:     []string{"LCL"},
			Label:       "Log Code Locations",
			Description: "whether to print code locations in logs",
			Documentation: strings.TrimSpace(strings.TrimSpace(`
Toggles on and off the printing of code locations in logs.
`)),
			Default: "true",
		}, func(o *toggle.Opt) (err error) {
			log2.CodeLoc = o.Value().Bool()
			return
		}),

		"LogLevel": text.New(meta.Data{
			Aliases: []string{"LL"},
			Label:   "Log Level",
			Description: "Level of logging to print: [ " + log2.LvlStr.String() +
				" ]",
			Documentation: strings.TrimSpace(`
Log levels are values in ascending order with the following names:	

	` + log2.LvlStr.String() + `

The level set in this configuration item defines the limit in ascending order
of what log level printers will output. Default is 'info' which means debug and 
trace log statements will not print. 
`),
			Default: log2.GetLevelName(log2.Info),
		}, func(o *text.Opt) (err error) {
			v := strings.TrimSpace(o.Value().Text())
			found := false
			lvl := log2.Info
			for i := range log2.LvlStr {
				ll := log2.GetLevelName(i)
				if util.Norm(v) == strings.TrimSpace(ll) {
					lvl = i
					found = true
				}
			}
			if !found {
				err = fmt.Errorf("log level value %s not valid from %v",
					v, log2.LvlStr)
				_ = o.FromString(log2.GetLevelName(lvl))
			}
			log2.SetLogLevel(lvl)
			return
		}),

		"LogFilePath": text.New(meta.Data{
			Aliases:     Tags("LFP"),
			Label:       "Log To File",
			Description: "Write logs to the specified file",
			Documentation: strings.TrimSpace(`
Sets the path of the file to write logs to.
`),
			Default: filepath.Join(defaultDataDir, "log.txt"),
		}, func(o *text.Opt) (err error) {
			err = log2.SetLogFilePath(o.Expanded())
			return
		}, text.NormalizeFilesystemPath(abs, appName)),

		"LogToFile": toggle.New(meta.Data{
			Aliases:     Tags("LTF"),
			Label:       "Log To File",
			Description: "Enable writing of logs",
			Documentation: strings.TrimSpace(`
Enables the writing of logs to the file path defined in LogFilePath.
`),
			Default: "false",
		}, func(o *toggle.Opt) (err error) {
			if o.Value().Bool() {
				log.T.Ln("starting log file writing")
				err = log2.StartLogToFile()
			} else {
				err = log2.StopLogToFile()
				log.T.Ln("stopped log file writing")
			}
			log.E.Chk(err)
			return
		}),
	}
	for i := range options {
		in[i] = options[i]
	}
}

// Init sets up a Command to be ready to use. Puts the reverse paths into the
// tree structure, puts sane defaults into command launchers, runs the hooks on
// all the defined configuration values, and sets the paths on each Command and
// Option so that they can be directly interrogated for their location.
func Init(c *Command, p path.Path) (cmd *Command, err error) {
	if c.Parent != nil {
		log.T.Ln("backlinking children of", c.Parent.Name)
	}
	if c.Entrypoint == nil {
		c.Entrypoint = NoOp
	}
	if p == nil {
		p = path.Path{c.Name}
	}
	c.Path = p // .Parent()
	for i := range c.Configs {
		c.Configs[i].SetPath(p)
	}
	for i := range c.Commands {
		c.Commands[i].Parent = c
		c.Commands[i].Path = p.Child(c.Commands[i].Name)
		_, _ = Init(c.Commands[i], p.Child(c.Commands[i].Name))
	}
	c.ForEach(func(cmd *Command, _ int) bool {
		for i := range cmd.Configs {
			err = cmd.Configs[i].RunHooks()
			if log.E.Chk(err) {
				return false
			}
		}
		return true
	}, 0, 0, c)
	return c, err
}

// GetOpt returns the option at a requested path
func (c *Command) GetOpt(path path.Path) (o config.Option) {
	p := make([]string, len(path))
	for i := range path {
		p[i] = path[i]
	}
	switch {
	case len(p) < 1:
		// not found
		return
	case len(p) > 2:
		// search subcommands
		for i := range c.Commands {
			if util.Norm(c.Commands[i].Name) == util.Norm(p[1]) {
				return c.Commands[i].GetOpt(p[1:])
			}
		}
	case len(p) == 2:
		// check name matches path, search for config item
		if util.Norm(c.Name) == util.Norm(p[0]) {
			for i := range c.Configs {
				if util.Norm(i) == util.Norm(p[1]) {
					return c.Configs[i]
				}
			}
		}
	}
	return nil
}

func (c *Command) GetCommand(p string) (o *Command) {
	pp := strings.Split(p, " ")
	path := path.Path(pp)
	// log.I.F("%v == %v", path, c.Path)
	if path.Equal(c.Path) {
		// log.I.Ln("found", c.Path)
		return c
	}
	for i := range c.Commands {
		// log.I.Ln(c.Commands[i].Path)
		o = c.Commands[i].GetCommand(p)
		if o != nil {
			return
		}
	}
	return
}

// ForEach runs a closure on every node in the Commands tree, stopping if the
// closure returns false
func (c *Command) ForEach(cl func(*Command, int) bool, hereDepth,
	hereDist int, cmd *Command) (ocl func(*Command, int) bool, depth,
	dist int, cm *Command) {
	ocl = cl
	cm = cmd
	if hereDepth == 0 {
		if !ocl(cm, hereDepth) {
			return
		}
	}
	depth = hereDepth + 1
	log.T.Ln(path.GetIndent(depth)+"->", depth)
	dist = hereDist
	for i := range c.Commands {
		log.T.Ln(path.GetIndent(depth)+"walking", c.Commands[i].Name, depth,
			dist)
		if !cl(c.Commands[i], depth) {
			return
		}
		dist++
		ocl, depth, dist, cm = c.Commands[i].ForEach(
			cl,
			depth,
			dist,
			cm,
		)
	}
	log.T.Ln(path.GetIndent(hereDepth)+"<-", hereDepth)
	depth--
	return
}
