package main

import (
	"fmt"
	"github.com/gizak/termui"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"unicode/utf8"
)

type Cmd struct {
	Cmd  string `yaml:"cmd"`
	Name string `yaml:"name"`
}

type SearchObj struct {
	MatchedIndexList []int
	SelfIndexInList  int
	QueryStr         string
	CommandSize      int
	SearchTitle      string
	SearchMode       bool
}

func (so *SearchObj) Reset() {
	so.MatchedIndexList = []int{}
	so.SelfIndexInList = 0
	so.QueryStr = ""
}

func (so *SearchObj) Title() string {
	if len(so.MatchedIndexList) > 0 {
		return fmt.Sprintf("%s%s     [共匹配到%d个,第%d个,按C-n/C-p导航] 按ESC退出搜索", so.SearchTitle, so.QueryStr, len(so.MatchedIndexList), so.SelfIndexInList+1)
	} else {
		return fmt.Sprintf("%s%s     按ESC退出搜索", so.SearchTitle, so.QueryStr)
	}
}
func (so *SearchObj) Next(current int) int {
	if so.SelfIndexInList >= 0 && so.SelfIndexInList < len(so.MatchedIndexList) {
		so.SelfIndexInList = (so.SelfIndexInList + 1) % len(so.MatchedIndexList)
		return so.MatchedIndexList[so.SelfIndexInList] - current
	}
	return 0
}

func (so *SearchObj) Offset(current int) int {
	if so.SelfIndexInList >= 0 && so.SelfIndexInList < len(so.MatchedIndexList) {
		return so.MatchedIndexList[so.SelfIndexInList] - current
	}
	return 0
}

func (so *SearchObj) Prev(current int) int {
	if so.SelfIndexInList >= 0 && so.SelfIndexInList < len(so.MatchedIndexList) {
		so.SelfIndexInList = (so.SelfIndexInList - 1 + len(so.MatchedIndexList)) % len(so.MatchedIndexList)
		return so.MatchedIndexList[so.SelfIndexInList] - current
	}
	return 0
}

var MaxLine int = 20

var configFile string = os.Getenv("HOME") + "/.relay.conf"

var commands []Cmd
var searchObj = &SearchObj{}
var currentIndex int = 0
var exitNow bool = false

func main() {
	commands = loadCommands()
	if len(commands) == 0 {
		fmt.Println("无主机配置")
		os.Exit(1)
	}
	drawUI()
	if !exitNow {
		runCommand(commands[currentIndex].Cmd)
	}
	os.Exit(0)
}

func drawUI() {
	searchObj.SearchMode = false
	searchObj.CommandSize = len(commands)
	searchObj.SearchTitle = "查找主机: "
	origTitle := "选择登录的主机 Help:(1: <TAB/j/k>进行选择 2: <C-d/C-u/g/G>翻页/第一行/最后一行 3: </>搜索 4: Enter确认 5: <ESC/q/C-c>退出)"
	err := termui.Init()
	if err != nil {
		panic(err)
	}
	defer termui.Close()
	// set max line
	if MaxLine < termui.TermHeight()-3 {
		MaxLine = termui.TermHeight() - 3
	}

	strs := formatCommands(commands, currentIndex)

	ls := termui.NewList()
	ls.Items = strs
	ls.ItemFgColor = termui.ColorYellow
	ls.BorderLabel = origTitle
	ls.Height = termui.TermHeight()
	termui.Body.AddRows(termui.NewRow(termui.NewCol(12, 0, ls)))

	termui.Body.Align()

	termui.Render(termui.Body)

	repaint := func(offset int) {
		nIndex := offset + currentIndex
		if nIndex < 0 {
			nIndex += len(commands)
		}
		currentIndex = nIndex % len(commands)
		ls.Items = formatCommands(commands, currentIndex)
		termui.Render(termui.Body)
	}

	doSearch := func() int {
		searchObj.MatchedIndexList = []int{}
		searchObj.SelfIndexInList = 0
		if searchObj.QueryStr != "" {
			for i, c := range commands {
				if strings.Contains(strings.ToLower(c.Name), strings.ToLower(searchObj.QueryStr)) {
					searchObj.MatchedIndexList = append(searchObj.MatchedIndexList, i)
				}
			}
		}
		ls.BorderLabel = searchObj.Title()
		return searchObj.Offset(currentIndex)
	}

	appendQuery := func(qs string) {
		searchObj.QueryStr += qs
		repaint(doSearch())
	}

	termui.Handle("/sys/kbd/<enter>", func(termui.Event) {
		termui.StopLoop()
	})
	termui.Handle("/sys/kbd/<escape>", func(termui.Event) {
		if searchObj.SearchMode {
			searchObj.SearchMode = false
			searchObj.Reset()
			ls.BorderLabel = origTitle
			repaint(0)
		} else {
			termui.StopLoop()
			exitNow = true
		}
	})
	termui.Handle("/sys/kbd/q", func(termui.Event) {
		if !searchObj.SearchMode {
			termui.StopLoop()
			exitNow = true
		} else {
			appendQuery("q")
		}
	})
	termui.Handle("/sys/kbd/C-c", func(termui.Event) {
		termui.StopLoop()
		exitNow = true
	})
	termui.Handle("/sys/kbd/<tab>", func(termui.Event) {
		if !searchObj.SearchMode {
			repaint(1)
		}
	})
	termui.Handle("/sys/kbd/C-n", func(termui.Event) {
		if searchObj.SearchMode {
			offset := searchObj.Next(currentIndex)
			ls.BorderLabel = searchObj.Title()
			repaint(offset)
		} else {
			repaint(1)
		}
	})
	termui.Handle("/sys/kbd/j", func(termui.Event) {
		if !searchObj.SearchMode {
			repaint(1)
		} else {
			appendQuery("j")
		}
	})
	termui.Handle("/sys/kbd/C-p", func(termui.Event) {
		if searchObj.SearchMode {
			offset := searchObj.Prev(currentIndex)
			ls.BorderLabel = searchObj.Title()
			repaint(offset)
		} else {
			repaint(-1)
		}
	})
	termui.Handle("/sys/kbd/k", func(termui.Event) {
		if !searchObj.SearchMode {
			repaint(-1)
		} else {
			appendQuery("k")
		}
	})
	termui.Handle("/sys/kbd/C-d", func(termui.Event) {
		if !searchObj.SearchMode {
			repaint(10)
		}
	})
	termui.Handle("/sys/kbd/C-u", func(termui.Event) {
		if !searchObj.SearchMode {
			repaint(-10)
		}
	})
	termui.Handle("/sys/kbd/G", func(termui.Event) {
		if !searchObj.SearchMode {
			repaint(-currentIndex - 1)
		} else {
			appendQuery("G")
		}
	})
	termui.Handle("/sys/kbd/g", func(termui.Event) {
		if !searchObj.SearchMode {
			repaint(-currentIndex)
		} else {
			appendQuery("g")
		}
	})
	termui.Handle("/sys/wnd/resize", func(termui.Event) {
		repaint(0)
	})

	termui.Handle("/sys/kbd", func(evt termui.Event) {
		kb, ok := evt.Data.(termui.EvtKbd)
		if !ok {
			return
		}
		if kb.KeyStr == "/" && !searchObj.SearchMode {
			searchObj.SearchMode = true
			searchObj.Reset()
			ls.BorderLabel = searchObj.Title()
			repaint(0)
		} else if searchObj.SearchMode {
			if kb.KeyStr == "C-8" {
				// delete char
				_, size := utf8.DecodeLastRuneInString(searchObj.QueryStr)
				searchObj.QueryStr = searchObj.QueryStr[:len(searchObj.QueryStr)-size]
				repaint(doSearch())
			} else if kb.KeyStr == "<space>" {
				appendQuery(" ")
			} else {
				matched, _ := regexp.MatchString(`<.+>|C\-[^c]`, kb.KeyStr)
				if !matched {
					appendQuery(kb.KeyStr)
				}
			}
		}
	})
	termui.Loop()
}
func formatCommands(commands []Cmd, index int) []string {
	matchedMap := make(map[int]int)
	for i, j := range searchObj.MatchedIndexList {
		matchedMap[j] = i
	}
	var strs []string
	start := index - MaxLine + 1
	if start < 0 {
		start = 0
	}
	end := start + MaxLine - 1
	if searchObj.SearchMode {
		start = searchObj.SelfIndexInList - MaxLine + 1
		if start < 0 {
			start = 0
		}
		end = start + MaxLine - 1
	}
	for i, c := range commands {
		j, ok := matchedMap[i]
		if searchObj.SearchMode {
			if !ok || (j < start || j > end) {
				continue
			}
		} else if i < start || i > end {
			continue
		}
		if !searchObj.SearchMode || (searchObj.SearchMode && ok) {
			if i == index {
				strs = append(strs, fmt.Sprintf("[%v] [%s](fg-blue,bg-green)", i+1, c.Name))
			} else {
				strs = append(strs, fmt.Sprintf("[%v] %s", i+1, c.Name))
			}
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
