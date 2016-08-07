package main

import (
	"fmt"
	"github.com/gizak/termui"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"os/exec"
)

type Cmd struct {
	Cmd  string `yaml:"cmd"`
	Name string `yaml:"name"`
}

var MaxLine int = 20

var configFile string = os.Getenv("HOME") + "/.relay.conf"

var commands []Cmd
var currentIndex int = 0

func main() {
	commands = loadCommands()
	if len(commands) == 0 {
		fmt.Println("无主机配置")
		os.Exit(1)
	}
	err := termui.Init()
	if err != nil {
		panic(err)
	}
	defer termui.Close()

	strs := formatCommands(commands, currentIndex)

	ls := termui.NewList()
	ls.Items = strs
	ls.ItemFgColor = termui.ColorYellow
	ls.BorderLabel = "选择登录的主机 Help:(1: <TAB/C-n/C-p/j/k>进行选择 2: <C-d/C-u/g/G>翻页/第一行/最后一行 3: Enter确认 4: <q/C-c>退出)"
	ls.Height = 600
	// build layout
	termui.Body.AddRows(termui.NewRow(termui.NewCol(12, 0, ls)))

	// calculate layout
	termui.Body.Align()

	termui.Render(termui.Body)

	//	termui.Render(ls)
	repaint := func(offset int) {
		nIndex := offset + currentIndex
		if nIndex < 0 {
			nIndex += len(commands)
		}
		currentIndex = nIndex % len(commands)
		ls.Items = formatCommands(commands, currentIndex)
		termui.Render(termui.Body)
	}

	exitNow := false
	termui.Handle("/sys/kbd/<enter>", func(termui.Event) {
		termui.StopLoop()
	})
	termui.Handle("/sys/kbd/q", func(termui.Event) {
		termui.StopLoop()
		exitNow = true
	})
	termui.Handle("/sys/kbd/C-c", func(termui.Event) {
		termui.StopLoop()
		exitNow = true
	})
	termui.Handle("/sys/kbd/<tab>", func(termui.Event) {
		repaint(1)
	})
	termui.Handle("/sys/kbd/C-n", func(termui.Event) {
		repaint(1)
	})
	termui.Handle("/sys/kbd/j", func(termui.Event) {
		repaint(1)
	})
	termui.Handle("/sys/kbd/C-p", func(termui.Event) {
		repaint(-1)
	})
	termui.Handle("/sys/kbd/k", func(termui.Event) {
		repaint(-1)
	})
	termui.Handle("/sys/kbd/C-d", func(termui.Event) {
		repaint(10)
	})
	termui.Handle("/sys/kbd/C-u", func(termui.Event) {
		repaint(-10)
	})
	termui.Handle("/sys/kbd/G", func(termui.Event) {
		repaint(-currentIndex - 1)
	})
	termui.Handle("/sys/kbd/g", func(termui.Event) {
		repaint(-currentIndex)
	})
	termui.Loop()
	termui.Close()
	if !exitNow {
		runCommand(commands[currentIndex].Cmd)
	}
	os.Exit(0)
}

func formatCommands(commands []Cmd, index int) []string {
	var strs []string
	start := index - MaxLine + 1
	if start < 0 {
		start = 0
	}
	end := start + MaxLine - 1
	for i, c := range commands {
		if i < start || i > end {
			continue
		}
		if i == index {
			strs = append(strs, fmt.Sprintf("[%v] [%s](fg-blue,bg-green)", i+1, c.Name))
		} else {
			strs = append(strs, fmt.Sprintf("[%v] %s", i+1, c.Name))
		}
	}
	return strs
}

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

func runCommand(cmdstr string) {
	cmd := exec.Command("bash", "-c", cmdstr)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Run()
}
