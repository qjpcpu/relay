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
}

// initial value for max lines per screen
var MaxLine int = 20

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
	// should be: relay [alias shortcut/last]
	if len(os.Args) >= 2 && os.Args[1] != "" {
		// relay last: run the latest command directly
		if cache, err := loadCache(); err == nil && os.Args[1] == "last" && cache.LastIndex < len(commands) {
			shortcut = true
			currentIndex = cache.LastIndex
		} else {
			// relay alias: run the command searched by alias
			for i, cmd := range commands {
				if cmd.Alias == os.Args[1] {
					shortcut = true
					currentIndex = i
				}
			}
		}
	}
	// if no shortcut specify, show selection UI
	if !shortcut {
		drawUI()
	}
	// if user press q/C-c,exit now; else run the command selected.
	if !exitNow {
		fmt.Printf("执行命令: \033[1;33m%s\033[0m\n\033[0;32m%s\033[0m\n", commands[currentIndex].Name, commands[currentIndex].Cmd)
		// cache the comand as lastest command
		cache := Cache{LastIndex: currentIndex}
		saveCache(cache)
		// record command history, write the command to shell history file, this is the simplest way I know, maybe someone would advise me better way
		changeCliHistory(os.Args[0], commands[currentIndex].Cmd)
		// run the command selected
		execCommand(commands[currentIndex].Cmd)
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
