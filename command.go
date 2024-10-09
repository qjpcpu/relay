package main

import (
	"bufio"
	"crypto/md5"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"reflect"
	"syscall"

	"regexp"
	"strings"

	"github.com/hoisie/mustache"
	"github.com/qjpcpu/common.v2/cli"
	"github.com/qjpcpu/common.v2/stringutil"
	"gopkg.in/yaml.v2"
)

type OptionItem struct {
	Desc string `json:"desc"`
	Val  string `json:"val"`
}

// relay command definition
type Cmd struct {
	// command string
	Cmd string `yaml:"cmd"`
	// commnad name,just for display
	Name string `yaml:"name"`
	// command shortcut, fast access
	Alias string `yaml:"alias"`
	// values
	Options     map[string][]OptionItem `yaml:"-"`
	OptionsRaw  map[string]interface{}  `yaml:"options" json:"-"`
	NeedConfirm bool                    `yaml:"confirm"`
	Defaults    map[string]string       `yaml:"defaults"`
	// real command
	RealCommand string
}

// load command from config file
func loadCommands(ctx *context) ([]Cmd, error) {
	var commands []Cmd

	if data, err := ioutil.ReadFile(ctx.getConfigFile()); err != nil {
		commands = bootstrapCommands(ctx)
	} else if err = yaml.Unmarshal(data, &commands); err != nil {
		fmt.Printf("fail to parse %s %v\n", ctx.getConfigFile(), err)
		return nil, err
	}
	if len(commands) == 0 {
		fmt.Println("no command list")
		return nil, errors.New("no command list")
	}
	return formatCommandList(commands), nil
}

func formatCommandList(commands []Cmd) []Cmd {
	for i, cmd := range commands {
		cmd.RealCommand = cmd.Cmd
		cmd.Options = make(map[string][]OptionItem)
		for name, val := range cmd.OptionsRaw {
			if val == nil {
				continue
			} else if arr, ok := val.([]interface{}); !ok || len(arr) == 0 {
				continue
			} else {
				for _, v := range arr {
					if reflect.TypeOf(v).Kind() == reflect.Map {
						vMap := reflect.ValueOf(v)
						oi := OptionItem{}
						for _, key := range vMap.MapKeys() {
							k, v := fmt.Sprint(key.Interface()), fmt.Sprint(vMap.MapIndex(key).Interface())
							if k == OptionDesc {
								oi.Desc = v
							} else if k == OptionVal {
								oi.Val = v
							}
						}
						cmd.Options[name] = append(cmd.Options[name], oi)
					} else {
						cmd.Options[name] = append(cmd.Options[name], OptionItem{Val: fmt.Sprint(v)})
					}
				}
			}
		}
		cmd.OptionsRaw = nil
		commands[i] = cmd
	}
	return commands
}

func findCommandByAlias(ctx *context, commands []Cmd) (index int, ok bool) {
	alias := ctx.getAlias()
	// relay alias: run the command searched by alias
	if !stringutil.IsBlankStr(alias) && alias != "!" && alias != "@" {
		for i, cmd := range commands {
			if cmd.Alias == alias {
				ctx.MarkAlias()
				ok = true
				index = i
				return
			}
		}
	}
	return
}

func confirmComand(ctx *context, cmd *Cmd) {
	if cmd.NeedConfirm {
		fmt.Println("Press Enter to continue...")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
	}
}

// exec comand
func execCommand(ctx *context, cmdstr string) {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "bash"
	}
	binary, lookErr := exec.LookPath(shell)
	if lookErr != nil {
		panic(lookErr)
	}
	args := []string{binary, "-i", "-c", cmdstr}
	env := os.Environ()
	execErr := syscall.Exec(binary, args, env)
	if execErr != nil {
		panic(execErr)
	}
}

func (c Cmd) GetName() string {
	return c.Name
}

func (c Cmd) md5() string {
	return fmt.Sprintf("%x", md5.Sum([]byte(c.Cmd)))
}

func (c Cmd) Equals(c1 Cmd) bool {
	return c.Cmd == c1.Cmd && c.Name == c1.Cmd && c.RealCommand == c1.RealCommand
}

func commands2Items(cs []Cmd) []string {
	var list []string
	for _, c := range cs {
		list = append(list, c.Name)
	}
	return list
}

func commands2Hints(cs []Cmd) []string {
	var list []string
	for _, c := range cs {
		list = append(list, c.Cmd)
	}
	return list
}

func bootstrapCommands(ctx *context) []Cmd {
	data := []byte(`- name: edit ~/.relay.conf
  cmd: vim ~/.relay.conf
- name: login dev server on aws
  cmd: ssh root@8.8.8.8
- name: view system info
  cmd: top`)
	var cmds []Cmd
	yaml.Unmarshal(data, &cmds)
	ioutil.WriteFile(ctx.getConfigFile(), data, 0644)
	return cmds
}

func (cmd *Cmd) Variables() []string {
	re := regexp.MustCompile(`{{\s*([a-zA-Z_0-9]+)\s*}}`)
	matchers := re.FindAllString(cmd.Cmd, -1)
	var vars []string
	for _, m := range matchers {
		v := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(m, `{{`), `}}`))
		if v != "" {
			vars = append(vars, v)
		}
	}
	return vars
}

func (cmd *Cmd) Populate(data map[string]string) {
	tmpl, err := mustache.ParseString(cmd.Cmd)
	if err != nil {
		return
	}
	cmd.RealCommand = tmpl.Render(data)
}

func populateCommand(ctx *context, cmd *Cmd) (err error) {
	params := make(map[string]string)
	variables := cmd.Variables()
	if len(variables) == 0 {
		return
	}

	prefillArguments := ctx.ExtraArguments()
	for i, v := range variables {
		if _, ok := params[v]; ok {
			continue
		}
		if i < len(prefillArguments) {
			params[v] = prefillArguments[i]
			continue
		}
		if cmd.Defaults != nil && cmd.Defaults[v] != "" {
			params[v] = cmd.Defaults[v]
			continue
		}
		var items, hints []string
		for _, value := range cmd.Options[v] {
			items = append(items, value.Val)
			hints = append(hints, fmt.Sprintf("%s/%s", value.Desc, value.Val))
		}
		idx := cli.SelectWithSearch(v, hints)
		if idx < 0 {
			return dummyErr
		}
		params[v] = strings.TrimSpace(items[idx])
	}
	cmd.Populate(params)
	return
}

func optionItemToSuggestions(options []OptionItem) (suggestions []cli.Suggest) {
	for i, opt := range options {
		if !stringutil.IsBlankStr(opt.Val) {
			desc := PromptTypeDefault
			if !stringutil.IsBlankStr(opt.Desc) {
				desc = opt.Desc
			}
			suggestions = append(suggestions, cli.Suggest{
				Text: options[i].Val,
				Desc: desc,
			})
		}
	}
	return
}
