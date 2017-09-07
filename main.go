package main

import (
	"fmt"
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
	// real command
	RealCommand string
}

// default configuration files[format:yaml]
var configFile string = os.Getenv("HOME") + "/.relay.conf"

// cache relay history
var cacheFile string = os.Getenv("HOME") + "/.relay_cache"

// all commands loaded
var commands []Cmd

// cursor to current command
var currentIndex int = 0

// exit relay right now
var exitNow bool = false

func main() {
	commands = loadCommands()
	if len(commands) == 0 {
		fmt.Println("无主机配置")
		os.Exit(1)
	}
	shortcut := false
	populateData := make(map[string]string)
	cache, err := loadCache()
	// should be: relay [alias shortcut/!/!!]
	if len(os.Args) >= 2 && os.Args[1] != "" {
		for loop := true; loop; loop = false {
			// relay !: run the latest command directly
			if err == nil && os.Args[1] == "!" && cache.LastIndex < len(commands) {
				shortcut = true
				currentIndex = cache.LastIndex
				populateData = cache.Data
				break
			}
			// relay @: run from history
			if err == nil && os.Args[1] == "@" && len(cache.History) > 0 {
				history := make([]string, len(cache.History))
				for i, c := range cache.History {
					history[len(cache.History)-i-1] = c
				}
				selects := &SelectList{
					SelectedIndex: currentIndex,
					Items:         history,
					SelectNothing: false,
				}
				selects.DrawUI()
				if !selects.SelectNothing {
					execCommand(history[selects.SelectedIndex])
				}
				os.Exit(0)
				break
			}
			// relay alias: run the command searched by alias
			for i, cmd := range commands {
				if cmd.Alias == os.Args[1] {
					shortcut = true
					currentIndex = i
				}
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
	// if user press q/C-c,exit now; else run the command selected.
	if !selects.SelectNothing {
		currentIndex = selects.SelectedIndex
		// populate command variables if exists
		if vlen := len(commands[currentIndex].Variables()); vlen == 0 || len(populateData) > 0 {
			populateData = populateCommand(&commands[currentIndex], populateData)
			fmt.Printf("执行命令: \033[1;33m%s\033[0m\n\033[0;32m%s\033[0m\n", commands[currentIndex].Name, commands[currentIndex].RealCommand)
		} else {
			fmt.Printf("命令\033[1;33m%s\033[0m需要填充变量:\n", commands[currentIndex].Name)
			populateData = populateCommand(&commands[currentIndex], populateData)
			fmt.Printf("执行命令: \033[0;32m%s\033[0m\n", commands[currentIndex].RealCommand)
		}
		// cache the comand as lastest command
		cache = Cache{LastIndex: currentIndex, Data: populateData, History: cache.History}
		cache.History = append(cache.History, commands[currentIndex].RealCommand)
		saveCache(cache)
		// run the command selected
		execCommand(commands[currentIndex].RealCommand)
	}
}

// load command from config file
func loadCommands() []Cmd {
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		fmt.Printf("读取配置文件%s失败", configFile)
		os.Exit(1)
	}
	var commands []Cmd
	err = yaml.Unmarshal(data, &commands)
	if err != nil {
		fmt.Printf("解析配置文件%s失败", configFile)
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

func commands2Items(cs []Cmd) []string {
	var list []string
	for _, c := range cs {
		list = append(list, c.Name)
	}
	return list
}
