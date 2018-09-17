package main

import (
	"errors"
	"fmt"
	"github.com/urfave/cli"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"os/exec"
	"syscall"
)

// relay command definition
type Cmd struct {
	// command string
	Cmd string `yaml:"cmd"`
	// commnad name,just for display
	Name string `yaml:"name"`
	// command shortcut, fast access
	Alias string `yaml:"alias"`
	// default values
	Defaults map[string]string `yaml:"defaults"`
	// real command
	RealCommand string
}

// default configuration files[format:yaml]
var configFile string = os.Getenv("HOME") + "/.relay.conf"

// cache relay history
var cacheFile string = os.Getenv("HOME") + "/.relay_cache"

func main() {
	app := cli.NewApp()
	app.Name = "relay"
	app.Usage = "command relay station"
	app.Authors = []cli.Author{
		cli.Author{
			Name:  "JasonQu",
			Email: "qjpcpu@gmail.com",
		},
	}
	app.UsageText = "relay [global options] [command alias] [arguments...]"
	app.HideHelp = true
	app.HideVersion = true
	app.EnableBashCompletion = true
	app.BashComplete = func(c *cli.Context) {
		// This will complete if no args are passed
		if c.NArg() > 0 {
			return
		}
		configFile = c.GlobalString("c")
		commands := loadCommands()
		for _, c := range commands {
			if c.Alias != "" {
				fmt.Println(c.Alias)
			}
		}
	}
	app.Before = func(c *cli.Context) error {
		configFile = c.GlobalString("c")
		return nil
	}
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "c",
			Usage: "specify config file",
			Value: os.Getenv("HOME") + "/.relay.conf",
		},
		cli.BoolFlag{
			Name:  "help, h",
			Usage: "show help",
		},
	}
	app.Action = func(c *cli.Context) error {
		if c.GlobalBool("help") {
			cli.ShowAppHelp(c)
			return nil
		}
		return runRelayCommand(c)
	}
	app.Commands = []cli.Command{
		{
			Name:  "!",
			Usage: "run last command",
			Action: func(c *cli.Context) error {
				return runLastCommand(c)
			},
		},
		{
			Name:  "@",
			Usage: "show relay history",
			Action: func(c *cli.Context) error {
				return runHistoryCommand(c)
			},
		},
	}
	app.Run(os.Args)
}

func runLastCommand(c *cli.Context) error {
	cache, err := loadCache()
	if err == nil && len(cache.History) > 0 {
		cmd := cache.History[len(cache.History)-1]
		cache.AppendHistory(cmd)
		saveCache(cache)
		fmt.Printf("Execute command: \033[1;33m%s\033[0m\n\033[0;32m%s\033[0m\n", cmd.Name, cmd.RealCommand)
		execCommand(cmd.RealCommand)
	} else {
		return runRelayCommand(c)
	}
	return nil
}

func runRelayCommand(c *cli.Context) error {
	shortcut := false
	alias := c.Args().Get(0)
	commands := loadCommands()
	currentIndex := 0
	if len(commands) == 0 {
		fmt.Println("no command list")
		return errors.New("no command list")
	}
	// relay alias: run the command searched by alias
	if alias != "" && alias != "!" && alias != "@" {
		for i, cmd := range commands {
			if cmd.Alias == alias {
				shortcut = true
				currentIndex = i
			}
		}
	}
	selects := &SelectList{
		SelectedIndex: currentIndex,
		Items:         commands2Items(commands),
		SelectNothing: false,
	}
	// if no shortcut specify, show selection UI
	if !shortcut {
		selects.DrawUI()
	}
	populateData := make(map[string]string)
	cache, _ := loadCache()
	// if user press q/C-c,exit now; else run the command selected.
	if !selects.SelectNothing {
		currentIndex := selects.SelectedIndex
		// fast run command like relay alias param1 param2 ...
		if vnames := commands[currentIndex].Variables(); len(vnames) == c.NArg()-1 {
			for i, vn := range vnames {
				populateData[vn] = c.Args().Get(i + 1)
			}
		}
		// populate command variables if exists
		if vlen := len(commands[currentIndex].Variables()); vlen == 0 || len(populateData) > 0 {
			populateData = populateCommand(&commands[currentIndex], populateData)
			fmt.Printf("Execute command: \033[1;33m%s\033[0m\n\033[0;32m%s\033[0m\n", commands[currentIndex].Name, commands[currentIndex].RealCommand)
		} else {
			fmt.Printf("Fill variables of command \033[1;33m%s\033[0m:\n", commands[currentIndex].Name)
			populateData = populateCommand(&commands[currentIndex], populateData)
			fmt.Printf("Execute commnad: \033[0;32m%s\033[0m\n", commands[currentIndex].RealCommand)
		}
		// cache the comand as lastest command
		cache.AppendHistory(commands[currentIndex])
		saveCache(cache)
		// run the command selected
		execCommand(commands[currentIndex].RealCommand)
	}
	return nil
}

func runHistoryCommand(c *cli.Context) error {
	cache, err := loadCache()
	if err != nil {
		fmt.Println("no history for now")
		return err
	}
	if len(cache.History) == 0 {
		fmt.Println("no history for now")
		return nil
	}
	history := make([]string, len(cache.History))
	history_names := make([]string, len(cache.History))
	for i, c := range cache.History {
		history[len(cache.History)-i-1] = c.RealCommand
		history_names[len(cache.History)-i-1] = c.Name + ": " + c.RealCommand
	}
	selects := &SelectList{
		SelectedIndex: 0,
		Items:         history_names,
		SelectNothing: false,
	}
	selects.DrawUI()
	if !selects.SelectNothing {
		execCommand(history[selects.SelectedIndex])
	}
	return nil
}

// load command from config file
func loadCommands() []Cmd {
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		fmt.Printf("faild to load %s\n", configFile)
		os.Exit(1)
	}
	var commands []Cmd
	err = yaml.Unmarshal(data, &commands)
	if err != nil {
		fmt.Printf("fail to parse %s\n", configFile)
		os.Exit(1)
	}
	for i, cmd := range commands {
		commands[i].RealCommand = cmd.Cmd
	}
	return commands
}

// exec comand
func execCommand(cmdstr string) {
	binary, lookErr := exec.LookPath("bash")
	if lookErr != nil {
		panic(lookErr)
	}
	args := []string{"bash", "-c", cmdstr}
	env := os.Environ()
	execErr := syscall.Exec(binary, args, env)
	if execErr != nil {
		panic(execErr)
	}
}

func (c Cmd) GetName() string {
	return c.Name
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
