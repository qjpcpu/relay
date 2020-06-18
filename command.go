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

	"gopkg.in/yaml.v2"

	"github.com/hoisie/mustache"
	"github.com/qjpcpu/go-prompt"
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
	if !isStrBlank(alias) && alias != "!" && alias != "@" {
		for i, cmd := range commands {
			if cmd.Alias == alias {
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

func completerWithDefault(key string, options []OptionItem, optValHis []string) func(prompt.Document) []prompt.Suggest {
	return func(d prompt.Document) []prompt.Suggest {
		var suggestions []prompt.Suggest

		optValHis = removeDupElem(optValHis)

		for i, opt := range options {
			if !isStrBlank(opt.Val) {
				desc := PromptTypeDefault
				if !isStrBlank(opt.Desc) {
					desc = opt.Desc
				}
				suggestions = append(suggestions, prompt.Suggest{
					Text:        options[i].Val,
					Description: desc,
				})
			}
		}

		files, _ := ioutil.ReadDir(".")
		for _, file := range files {
			if strings.HasPrefix(file.Name(), ".") {
				continue
			}
			suggestions = append(suggestions, prompt.Suggest{
				Text:        file.Name(),
				Description: fileDesc(file),
			})
		}

		/* insert history */
		for i := 0; i < len(optValHis); i++ {
			val := optValHis[i]
			idx := -1
			for j, sg := range suggestions {
				if sg.Text == val {
					idx = j
					break
				}
			}
			if idx == -1 {
				suggestions = append([]prompt.Suggest{{
					Text:        val,
					Description: PromptTypeHistory,
				}}, suggestions...)
			} else if idx > 0 {
				tmp := suggestions[idx]
				for j := idx; j > 0; j-- {
					suggestions[j] = suggestions[j-1]
				}
				suggestions[0] = tmp
			}
		}

		return prompt.FilterContains(suggestions, d.GetWordBeforeCursor(), true)
	}
}

func populateCommand(cmd *Cmd, optHis map[string][]string) (params map[string]string, err error) {
	params = make(map[string]string)
	variables := cmd.Variables()
	if len(variables) == 0 {
		return
	}
	if optHis == nil {
		optHis = make(map[string][]string)
	}
	for _, v := range variables {
		if _, ok := params[v]; ok {
			continue
		}
		header := fmt.Sprintf("%s %s ", v, ParamInputHintSymbol)
		text, shouldExit := prompt.Input(
			header,
			completerWithDefault(v, cmd.Options[v], optHis[v]),
			prompt.OptionPrefixTextColor(prompt.Blue),
		)
		if shouldExit {
			return params, dummyErr
		}
		params[v] = strings.TrimSpace(text)
	}
	cmd.Populate(params)
	return
}
