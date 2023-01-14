package cmds

import (
	"os"
	"sort"
	"strings"

	"github.com/indra-labs/indra/pkg/opts/config"
	"github.com/indra-labs/indra/pkg/path"
)

type Env struct {
	Name path.Path
	Opt  config.Option
}

type Envs []Env

// ForEach runs a closure on each element of an Env item in an Envs and halts if
// one returns an error.
func (e Envs) ForEach(fn func(env string, opt config.Option) (err error)) (err error) {
	for i := range e {
		var name []string
		for j := range e[i].Name {
			name = append(name, strings.ToUpper(e[i].Name[j]))
		}
		err = fn(strings.Join(name, "_"), e[i].Opt)
		if err != nil {
			return
		}
	}
	return
}

// LoadFromEnvironment walks the Envs and reads the values found in the
// environment variables.
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

// Len returns the number items in an Envs. This is a sort.Sorter interface
// implementation.
func (e Envs) Len() int { return len(e) }

// Less compares two items and returns whether the first is lesser than the
// second. This is a sort.Sorter interface implementation.
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

// Swap switches the position of two elements of an Envs. This is a sort.Sorter
// interface implementation.
func (e Envs) Swap(i, j int) { e[i], e[j] = e[j], e[i] }

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
