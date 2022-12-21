package cmds

import (
	"encoding"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	integer "github.com/cybriq/proc/pkg/opts/Integer"
	"github.com/cybriq/proc/pkg/opts/duration"
	"github.com/cybriq/proc/pkg/opts/float"
	"github.com/cybriq/proc/pkg/opts/list"
	"github.com/cybriq/proc/pkg/opts/meta"
	"github.com/cybriq/proc/pkg/opts/text"
	"github.com/cybriq/proc/pkg/opts/toggle"
	path2 "github.com/cybriq/proc/pkg/path"
	"github.com/naoina/toml"
)

type Entry struct {
	path  path2.Path
	name  string
	value interface{}
}

func (e Entry) String() string {
	return fmt.Sprint(e.path, "/", e.name, "=", e.value)
}

type Entries []Entry

func (e Entries) Len() int {
	return len(e)
}

func (e Entries) Less(i, j int) bool {
	iPath, jPath := e[i].String(), e[j].String()
	return iPath < jPath
}

func (e Entries) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

func walk(p []string, v interface{}, in Entries) (o Entries) {
	o = append(o, in...)
	var parent []string
	for i := range p {
		parent = append(parent, p[i])
	}
	switch vv := v.(type) {
	case map[string]interface{}:
		for i := range vv {
			switch vvv := vv[i].(type) {
			case map[string]interface{}:
				o = walk(append(parent, i), vvv, o)
			default:
				o = append(o, Entry{
					path:  append(parent, i),
					name:  i,
					value: vv[i],
				})

			}
		}
	}
	return
}

var _ encoding.TextMarshaler = &Command{}

func (c *Command) MarshalText() (text []byte, err error) {
	c.ForEach(func(cmd *Command, depth int) bool {
		if cmd == nil {
			log.I.Ln("cmd empty")
			return true
		}
		cfgNames := make([]string, 0, len(cmd.Configs))
		for i := range cmd.Configs {
			cfgNames = append(cfgNames, i)
		}
		if len(cfgNames) < 1 {
			return true
		}
		if cmd.Name != "" {
			var cmdPath string
			current := cmd.Parent
			for current != nil {
				if current.Name != "" {
					cmdPath = current.Name + "."
				}
				current = current.Parent
			}
			text = append(text,
				[]byte("# "+cmdPath+cmd.Name+": "+cmd.Description+"\n")...)
			text = append(text,
				[]byte("["+cmdPath+cmd.Name+"]"+"\n\n")...)
		}
		sort.Strings(cfgNames)
		for _, i := range cfgNames {
			md := cmd.Configs[i].Meta()
			lq, rq := "", ""
			st := cmd.Configs[i].String()
			df := md.Default()
			switch cmd.Configs[i].Type() {
			case meta.Duration, meta.Text:
				lq, rq = "\"", "\""
			case meta.List:
				lq, rq = "[ \"", "\" ]"
				st = strings.ReplaceAll(st, ",", "\", \"")
				df = strings.ReplaceAll(df, ",", "\", \"")
				if st == "" {
					lq, rq = "[ ", "]"
				}
			}
			text = append(text,
				[]byte("# "+i+" - "+md.Description()+
					" - default: "+lq+df+rq+"\n")...)
			text = append(text,
				[]byte(i+" = "+lq+st+rq+"\n")...)
		}
		text = append(text, []byte("\n")...)
		return true
	}, 0, 0, c)
	return
}

var _ encoding.TextUnmarshaler = &Command{}

func (c *Command) UnmarshalText(t []byte) (err error) {
	var out interface{}
	err = toml.Unmarshal(t, &out)
	oo := walk([]string{}, out, []Entry{})
	sort.Sort(oo)
	for i := range oo {
		op := c.GetOpt(oo[i].path)
		if op != nil {
			switch op.Type() {
			case meta.Bool:
				v := op.(*toggle.Opt)
				nv, ok := oo[i].value.(bool)
				if ok {
					v.FromValue(nv)
				}
				log.T.Ln("setting value of", oo[i].path, "to", nv)
			case meta.Duration:
				v := op.(*duration.Opt)
				nv, ok := oo[i].value.(time.Duration)
				if ok {
					v.FromValue(nv)
				}
				log.T.Ln("setting value of", oo[i].path, "to", nv)
			case meta.Float:
				v := op.(*float.Opt)
				nv, ok := oo[i].value.(float64)
				if ok {
					v.FromValue(nv)
				}
				log.T.Ln("setting value of", oo[i].path, "to", nv)
			case meta.Integer:
				v := op.(*integer.Opt)
				nv, ok := oo[i].value.(int64)
				if ok {
					v.FromValue(nv)
				}
				log.T.Ln("setting value of", oo[i].path, "to", nv)
			case meta.List:
				v := op.(*list.Opt)
				nv, ok := oo[i].value.([]interface{})
				var strings []string
				for i := range nv {
					strings = append(strings, nv[i].(string))
				}
				if ok {
					v.FromValue(strings)
				}
				log.T.Ln("setting value of", oo[i].path, "to", nv)
			case meta.Text:
				v := op.(*text.Opt)
				nv, ok := oo[i].value.(string)
				if ok {
					v.FromValue(nv)
				}
				log.T.Ln("setting value of", oo[i].path, "to", nv)
			default:
				log.E.Ln("option type unknown:", oo[i].path, op.Type())
			}
		} else {
			log.E.Ln("option not found:", oo[i].path)
		}
	}
	return nil
}
func (c *Command) LoadConfig() (err error) {
	cfgFile := c.GetOpt(path2.Path{c.Name, "ConfigFile"})
	var file io.Reader
	if file, err = os.Open(cfgFile.Expanded()); err != nil {
		log.T.F("creating config file at path: '%s'", cfgFile.Expanded())
		// If no config found, create data dir and drop the default in place
		return c.SaveConfig()
	} else {
		var all []byte
		all, err = io.ReadAll(file)
		err = c.UnmarshalText(all)
		if log.E.Chk(err) {
			return
		}
	}
	return
}

func (c *Command) SaveConfig() (err error) {
	datadir := c.GetOpt(path2.Path{c.Name, "DataDir"})
	if err = os.MkdirAll(datadir.Expanded(), 0700); log.E.Chk(err) {
		return err
	}
	var f *os.File
	cfgFile := c.GetOpt(path2.From("pod123 configfile"))
	f, err = os.OpenFile(cfgFile.Expanded(), os.O_RDWR|os.O_CREATE, 0666)
	if log.E.Chk(err) {
		return
	}
	var cf []byte
	if cf, err = c.MarshalText(); log.E.Chk(err) {
		return
	}
	_, err = f.Write(cf)
	log.E.Chk(err)
	return
}
