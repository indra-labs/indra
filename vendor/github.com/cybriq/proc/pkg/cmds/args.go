package cmds

import (
	"fmt"
	"strings"

	"github.com/cybriq/proc/pkg/opts/meta"
	"github.com/cybriq/proc/pkg/util"
)

// ParseCLIArgs reads a command line argument slice (presumably from
// os.Args), identifies the command to run and
//
// Rules for constructing CLI args:
//
// - Commands are identified by name, and must appear in their hierarchic
//   order to invoke subcommands. They are matched as normalised to lower
//   case.
//
// - Options can be preceded by "--" or "-", and the full name, or the
//   alias, normalised to lower case for matching, and if there is an "="
//   after it, the value is after this, otherwise, the next element in the
//   args is the value, except booleans, which default to true unless set
//   to false. For these, the prefix "no" or similar indicates the
//   semantics of the true value.
//
// - Options only match when preceded by their relevant Command, except for
//   the root Command, and these options must precede any other command
//   options.
//
// - If no command is selected, the root Command.Default is selected. This
//   can optionally be used for subcommands as well, though it is unlikely
//   needed, if found, the Default of the tip of the Command branch
//   selected by the CLI if there is one, otherwise the Command itself.
func (c *Command) ParseCLIArgs(a []string) (run *Command, runArgs []string, err error) {
	args := make([]string, len(a))
	var cursor int
	for i := range a {
		if len(a[i]) > 0 {
			args[cursor] = a[i]
			cursor++
		}
	}
	var segments [][]string
	commands := Commands{c}
	var depth, last int
	var done bool
	current := c
	cursor = 0
	// First pass matches Command names in order to slice up the sections
	// where relevant config items will be found.
	for !done {
		for i := range current.Commands {
			if util.Norm(args[cursor]) == util.Norm(current.Commands[i].Name) {
				// the command to run is the last, so update each new command
				// found
				run = current.Commands[i]
				commands = append(commands, run)
				depth++
				current = current.Commands[i]
				segments = append(segments, args[last:cursor])
				last = cursor
				break
			}
		}
		cursor++
		// append the remainder to the last segment
		if cursor == len(args) {
			var tmp []string
			for _, item := range args[last:cursor] {
				if len(item) > 0 {
					tmp = append(tmp, item)
				}
			}
			segments = append(segments, tmp)
			done = true
		}
	}
	// The segments that have been cut from args will now provide the root level
	// command name, and all subsequent items until the next segment should be
	// names found in the configs map.
	for i := range segments {
		if len(segments[i]) > 0 {
			iArgs := segments[i][1:]
			cmd := commands[i]
			// log.D.Ln(commands[i].Name, "args", iArgs)
			// the final command can accept arbitrary arguments, that are passed
			// into the endrypoint
			runArgs = iArgs
			if util.Norm(commands[i].Name) == "help" {
				break
			}
			var cursor int
			for cursor < len(iArgs) {
				inc := 1
				arg := iArgs[cursor]
				if len(arg) == 0 {
					cursor++
					continue
				}
				log.D.Ln("evaluating", arg, iArgs[cursor:])
				if strings.HasPrefix(arg, "-") {
					arg = arg[1:]
					if strings.HasPrefix(arg, "-") {
						arg = arg[1:]
					}
					if strings.Contains(arg, "=") {
						log.D.Ln("value in arg", arg)
						split := strings.Split(arg, "=")
						if len(split) > 2 {
							split = append(split[:1], strings.Join(split[1:], "="))
						}
						for cfgName := range cmd.Configs {
							aliases := cmd.Configs[cfgName].Meta().Aliases()
							names := append(
								[]string{cfgName}, aliases...)
							for _, name := range names {
								if util.Norm(name) == util.Norm(split[0]) {
									log.D.F("assigning value '%s' to %s",
										split[1], split[0])
									err = cmd.Configs[cfgName].FromString(split[1])
									if log.E.Chk(err) {
										return
									}
								}
							}
						}
					} else {
						if len(iArgs) > cursor {
							found := false
							for cfgName := range cmd.Configs {
								aliases := cmd.Configs[cfgName].Meta().Aliases()
								names := append(
									[]string{cfgName}, aliases...)
								for _, name := range names {
									if util.Norm(name) == util.Norm(arg) {
										// check for booleans, which can only be
										// followed by true or false
										if cmd.Configs[cfgName].Type() == meta.Bool {
											err = cmd.Configs[cfgName].
												FromString(iArgs[cursor+1])
											// next value is not a truth value,
											// simply assign true and increment
											// only 1 to cursor
											if err != nil {
												log.D.Chk(err)
												found = true
												log.D.F("assigned value 'true' to %s",
													cfgName)
												break
											}
										}
										log.D.F("assigning value '%s' to %s",
											iArgs[cursor+1], cfgName)
										err = cmd.Configs[cfgName].
											FromString(iArgs[cursor+1])
										if log.E.Chk(err) {
											return
										}
										inc++
										found = true
									}
								}
							}
							if !found {
								err = fmt.Errorf(
									"option not found: '%s' context %v",
									arg, segments[i])
								return
							}
							// if this is the last arg, and it's bool, the
							// implied value is true
						} else if cmd.Configs[arg].Type() == meta.Bool {
							err = cmd.Configs[arg].FromString("true")
							if log.E.Chk(err) {
								return
							}
						} else {
							err = fmt.Errorf("argument '%s' missing value:"+
								"context %v", arg, iArgs)
							log.E.Chk(err)
							return
						}
					}
				} else {
					err = fmt.Errorf("argument %s missing '-', context %s, "+
						"most likely misspelled subcommand", arg, iArgs)
					log.E.Chk(err)
					return
				}
				cursor += inc
			}
		}
	}
	// if no Command was found, return the default. If there is no default, the
	// top level Command will be returned
	if len(c.Default) > 0 && len(segments) < 2 {
		run = c
		def := c.Default
		var lastFound int
		for i := range def {
			for _, sc := range run.Commands {
				if sc.Name == def[i] {
					lastFound = i
					run = sc
				}
			}
		}
		if lastFound != len(def)-1 {
			err = fmt.Errorf("default command %v not found at %s", c.Default,
				def)
		}
	}
	// log.D.F("will be executing command '%s' %s", run.Name, runArgs)

	return
}
