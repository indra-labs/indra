package cmds

import (
	"os"
	"sort"
	"strings"

	"github.com/cybriq/proc/pkg/opts/config"
	"github.com/cybriq/proc/pkg/path"
)

type Env struct {
	Name path.Path
	Opt  config.Option
}

type Envs []Env

func (e Envs) ForEach(fn func(env string, opt config.Option) (err error)) (err error) {
	for i := range e {
		var name []string
		for j := range e[i].Name {
			name = append(name, strings.ToUpper(e[i].Name[j]))
		}
		err = fn(strings.Join(name, "_"), nil)
		if err != nil {
			return
		}
	}
	return
}

func (e Envs) LoadFromEnvironment() (err error) {
	err = e.ForEach(func(env string, opt config.Option) (err error) {
		v, exists := os.LookupEnv(env)
		if exists {
			err = opt.FromString(v)
			if log.D.Chk(err) {
				return err
			}
		}
		return
	})
	return
}

func (e Envs) Len() int {
	return len(e)
}

func (e Envs) Less(i, j int) (res bool) {
	li, lj := len(e[i].Name), len(e[j].Name)
	if li < lj {
		return true
	}
	cursor := -1
	for {
		res = false
		cursor++
		if strings.Join(e[i].Name[:cursor], "_") <
			strings.Join(e[j].Name[:cursor], "_") {
			res = true
		}
		if cursor >= li || cursor >= lj {
			break
		}
	}
	return
}

func (e Envs) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

// GetEnvs walks a Command tree and returns a slice containing all environment
// variable names and the related config.Option.
func (c *Command) GetEnvs(path ...string) (envs Envs) {
	if path == nil {
		path = []string{c.Name}
	}
	for {
		for i := range c.Configs {
			envs = append(envs, Env{
				Name: append(path, i),
				Opt:  c.Configs[i],
			})
		}
		if len(c.Commands) > 0 {
			for i := range c.Commands {
				envs = append(envs,
					c.Commands[i].GetEnvs(
						append(path, c.Commands[i].Name)...)...,
				)
			}
		}
		break
	}
	sort.Sort(envs)
	return
}
