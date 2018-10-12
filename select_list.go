package main

import (
	"fmt"
	"github.com/gizak/termui"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

type SelectList struct {
	SelectedIndex int
	Items         []string
	shortItems    []string
	SelectNothing bool
}

var MaxLine = 20

var searchObj = &SearchObj{}

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

// render raw string as highlight format
func (so *SearchObj) Highlight(raw string, background bool) string {
	qs := strings.ToLower(so.QueryStr)
	raw1 := strings.ToLower(raw)
	if so.SearchMode && so.QueryStr != "" && FuzzyContains(raw1, qs) {
		start, matched := FuzzyIndex(raw1, qs)
		if start < 0 {
			if background {
				return fmt.Sprintf("[%s](fg-blue,bg-green)", raw)
			} else {
				return raw
			}
		}
		end := start + len(matched)
		raw1 = ""
		if start > 0 {
			raw1 += fmt.Sprintf("[%s](fg-magenta,bg-green)", raw[0:start])
		}
		raw1 += fmt.Sprintf("[%s](fg-white,fg-bold,bg-green)", raw[start:end])
		if end < len(raw) {
			raw1 += fmt.Sprintf("[%s](fg-magenta,bg-green)", raw[end:len(raw)])
		}
		if !background {
			raw1 = strings.Replace(raw1, ",bg-green", "", -1)
		} else {
			raw1 = strings.Replace(raw1, "fg-magenta", "fg-blue", -1)
		}
		return raw1
	} else {
		if background {
			return fmt.Sprintf("[%s](fg-blue,bg-green)", raw)
		} else {
			return raw
		}
	}
}

func (so *SearchObj) Title() string {
	if len(so.MatchedIndexList) > 0 {
		return fmt.Sprintf("%s%s     [Total %d items, current @%d, Navigation by C-n/C-p/down/up] press ESC exit search", so.SearchTitle, so.QueryStr, len(so.MatchedIndexList), so.SelfIndexInList+1)
	} else {
		return fmt.Sprintf("%s%s     Press ESC exit search", so.SearchTitle, so.QueryStr)
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

func (slist *SelectList) DrawUI() {
	for _, item := range slist.Items {
		slist.shortItems = append(slist.shortItems, item)
	}
	searchObj = &SearchObj{}
	searchObj.SearchMode = false
	searchObj.CommandSize = len(slist.Items)
	searchObj.SearchTitle = "Search: "

	origTitle := "Help:(1: <Enter>Confirm 2: </|C-s>Search 3: <ESC|q|C-c>Exit)"
	err := termui.Init()
	if err != nil {
		panic(err)
	}
	defer termui.Close()
	// set max line
	if MaxLine < termui.TermHeight()-3 {
		MaxLine = termui.TermHeight() - 3
	}

	strs := formatCommands(slist.shortItems, slist.SelectedIndex)

	ls := termui.NewList()
	ls.Items = strs
	ls.ItemFgColor = termui.ColorCyan
	ls.BorderLabel = origTitle
	ls.Height = termui.TermHeight()
	termui.Body.AddRows(termui.NewRow(termui.NewCol(12, 0, ls)))

	termui.Body.Align()

	termui.Render(termui.Body)

	repaint := func(offset int) {
		nIndex := offset + slist.SelectedIndex
		if nIndex < 0 {
			nIndex += len(slist.Items)
		}
		slist.SelectedIndex = nIndex % len(slist.Items)
		ls.Items = formatCommands(slist.shortItems, slist.SelectedIndex)
		termui.Render(termui.Body)
	}

	// term writer for fast select by number
	termWriter := NewTermWriter()
	termWriter.AddTerm("0gg", ``, "gg")
	termWriter.AddTerm("gg", `\d+`, "gg")
	termWriter.AddTerm("G", ``, "G")
	defer termWriter.Stop()
	go func() {
		for term := range termWriter.DataChan() {
			if !searchObj.SearchMode && term.IsMatched() {
				switch term.Name {
				case "0gg":
					repaint(-slist.SelectedIndex)
				case "gg":
					if idx, err := strconv.Atoi(strings.TrimSuffix(term.Text, "gg")); err == nil {
						repaint(idx - 1 - slist.SelectedIndex)
					}
				case "G":
					repaint(-slist.SelectedIndex - 1)

				}
			}
		}
	}()

	doSearch := func() int {
		searchObj.MatchedIndexList = []int{}
		searchObj.SelfIndexInList = 0
		if searchObj.QueryStr != "" {
			for i, c := range slist.Items {
				if FuzzyContains(strings.ToLower(c), strings.ToLower(searchObj.QueryStr)) {
					searchObj.MatchedIndexList = append(searchObj.MatchedIndexList, i)
				}
			}
		}
		ls.BorderLabel = searchObj.Title()
		return searchObj.Offset(slist.SelectedIndex)
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
			slist.SelectNothing = true
		}
	})
	termui.Handle("/sys/kbd/q", func(termui.Event) {
		if !searchObj.SearchMode {
			termui.StopLoop()
			slist.SelectNothing = true
		} else {
			appendQuery("q")
		}
	})
	termui.Handle("/sys/kbd/C-c", func(termui.Event) {
		termui.StopLoop()
		slist.SelectNothing = true
	})
	termui.Handle("/sys/kbd/<tab>", func(termui.Event) {
		if !searchObj.SearchMode {
			repaint(1)
		}
	})
	termui.Handle("/sys/kbd/C-n", func(termui.Event) {
		if searchObj.SearchMode {
			offset := searchObj.Next(slist.SelectedIndex)
			ls.BorderLabel = searchObj.Title()
			repaint(offset)
		} else {
			repaint(1)
		}
	})
	termui.Handle("/sys/kbd/<down>", func(termui.Event) {
		if searchObj.SearchMode {
			offset := searchObj.Next(slist.SelectedIndex)
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
			offset := searchObj.Prev(slist.SelectedIndex)
			ls.BorderLabel = searchObj.Title()
			repaint(offset)
		} else {
			repaint(-1)
		}
	})
	termui.Handle("/sys/kbd/<up>", func(termui.Event) {
		if searchObj.SearchMode {
			offset := searchObj.Prev(slist.SelectedIndex)
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
	// vim page down
	termui.Handle("/sys/kbd/C-d", func(termui.Event) {
		if !searchObj.SearchMode {
			repaint(10)
		}
	})
	// emacs page down
	termui.Handle("/sys/kbd/C-v", func(termui.Event) {
		if !searchObj.SearchMode {
			repaint(10)
		}
	})
	// vim page up
	termui.Handle("/sys/kbd/C-u", func(termui.Event) {
		if !searchObj.SearchMode {
			repaint(-10)
		}
	})
	// emacs page up
	termui.Handle("/sys/kbd/√", func(termui.Event) {
		if !searchObj.SearchMode {
			repaint(-10)
		}
	})
	termui.Handle("/sys/kbd/G", func(termui.Event) {
		if !searchObj.SearchMode {
			//repaint(-slist.SelectedIndex - 1)
			termWriter.Write([]byte("G"))
		} else {
			appendQuery("G")
		}
	})
	termui.Handle("/sys/kbd/g", func(termui.Event) {
		if !searchObj.SearchMode {
			//			repaint(-slist.SelectedIndex)
			termWriter.Write([]byte("g"))
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
		if !searchObj.SearchMode {
			switch {
			// enter search mode
			case kb.KeyStr == "/" || kb.KeyStr == "C-s":
				searchObj.SearchMode = true
				searchObj.Reset()
				ls.BorderLabel = searchObj.Title()
				repaint(0)
			case isNumber(kb.KeyStr):
				termWriter.Write([]byte(kb.KeyStr))
			}

		} else {
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

func isNumber(num string) bool {
	for i := 0; i <= 9; i++ {
		if strconv.Itoa(i) == num {
			return true
		}
	}
	return false
}
func arrayContains(arr []string, elem string) bool {
	if elem == "" {
		return false
	}
	for _, a := range arr {
		if a == elem {
			return true
		}
	}
	return false
}

// format command for UI display
func formatCommands(commands []string, index int) []string {
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
			fmtI := "%02d"
			if len(commands) > 100 {
				fmtI = "%03d"
			}
			if i == index {
				strs = append(strs, fmt.Sprintf("["+fmtI+"] %s", i+1, searchObj.Highlight(c, true)))
			} else {
				strs = append(strs, fmt.Sprintf("["+fmtI+"] %s", i+1, searchObj.Highlight(c, false)))
			}
		}
	}
	return strs
}
