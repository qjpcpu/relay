package main

import (
	"fmt"
	"github.com/gizak/termui"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

type selectListMode int

const (
	ModeNorm selectListMode = iota
	ModeSearch
)

type SelectList struct {
	// all items user can select from
	items []string
	// selected item index
	selectedIndex int
	// whether nothing selected
	selectNothing bool

	// max lines display in single screen
	maxLine int
	// original title
	title string
	// term writer, parse term input
	termWriter *TermWriter
	uilist     *termui.List
	mode       selectListMode
	search     *SearchObj
}

func NewSelectList(initialIndex int, items []string) *SelectList {
	// term writer for fast select by number
	termWriter := NewTermWriter()
	termWriter.AddTerm("0gg", ``, "gg")
	termWriter.AddTerm("gg", `\d+`, "gg")
	termWriter.AddTerm("G", ``, "G")
	return &SelectList{
		selectedIndex: initialIndex,
		items:         items,
		selectNothing: false,
		maxLine:       20,
		mode:          ModeNorm,
		termWriter:    termWriter,
		search: &SearchObj{
			SearchTitle: "Search: ",
		},
		title: "Help:(1: <Enter>Confirm 2: </|C-s>Search 3: <ESC|q|C-c>Exit)",
	}
}

// IsSelectNothing true: exit with nothing selected
func (sl *SelectList) IsSelectNothing() bool {
	return sl.selectNothing
}

// Selected return selected index
func (sl *SelectList) Selected() int {
	return sl.selectedIndex
}

// InSearchMode is in search mode
func (sl *SelectList) InSearchMode() bool {
	return sl.mode == ModeSearch
}

func (sl *SelectList) InNormMode() bool {
	return sl.mode == ModeNorm
}

// private methods
// move cursor by offset and repaint
func (sl *SelectList) repaint(offset int) {
	nIndex := offset + sl.selectedIndex
	if nIndex < 0 {
		nIndex += len(sl.items)
	}
	sl.selectedIndex = nIndex % len(sl.items)
	sl.uilist.Items = sl.formatCommands()
	termui.Render(termui.Body)
}

func (slist *SelectList) createUI() {
	ls := termui.NewList()
	ls.Items = slist.formatCommands()
	ls.ItemFgColor = termui.ColorCyan
	ls.BorderLabel = slist.title
	ls.Height = termui.TermHeight()
	termui.Body.AddRows(termui.NewRow(termui.NewCol(12, 0, ls)))

	slist.uilist = ls

	termui.Body.Align()

	termui.Render(termui.Body)
}

func (sl *SelectList) doSearch() int {
	sl.search.SearchResultsIndices = []int{}
	sl.search.SelectedResultIndex = 0
	if sl.search.QueryStr != "" {
		for i, c := range sl.items {
			if FuzzyContains(strings.ToLower(c), strings.ToLower(sl.search.QueryStr)) {
				sl.search.SearchResultsIndices = append(sl.search.SearchResultsIndices, i)
			}
		}
	}
	sl.uilist.BorderLabel = sl.search.Title()
	offset := sl.search.Offset(sl.selectedIndex)
	sl.repaint(offset)
	return offset
}

func (sl *SelectList) appendQuery(qs string) {
	sl.search.QueryStr += qs
	sl.doSearch()
}

func (sl *SelectList) reset() {
	sl.mode = ModeNorm
	sl.search.Reset()
	sl.uilist.BorderLabel = sl.title
}

func (slist *SelectList) handleKeyboardEvents() {
	termui.Handle("/sys/kbd/<enter>", func(termui.Event) {
		termui.StopLoop()
	})
	termui.Handle("/sys/kbd/<escape>", func(termui.Event) {
		if slist.InSearchMode() {
			slist.reset()
			slist.repaint(0)
		} else {
			termui.StopLoop()
			slist.selectNothing = true
		}
	})
	termui.Handle("/sys/kbd/q", func(termui.Event) {
		if slist.InNormMode() {
			termui.StopLoop()
			slist.selectNothing = true
		} else {
			slist.appendQuery("q")
		}
	})
	termui.Handle("/sys/kbd/C-c", func(termui.Event) {
		termui.StopLoop()
		slist.selectNothing = true
	})
	termui.Handle("/sys/kbd/<tab>", func(termui.Event) {
		if slist.InNormMode() {
			slist.repaint(1)
		} else {
			offset := slist.search.Next(slist.selectedIndex)
			slist.uilist.BorderLabel = slist.search.Title()
			slist.repaint(offset)
		}
	})
	termui.Handle("/sys/kbd/C-n", func(termui.Event) {
		if slist.InSearchMode() {
			offset := slist.search.Next(slist.selectedIndex)
			slist.uilist.BorderLabel = slist.search.Title()
			slist.repaint(offset)
		} else {
			slist.repaint(1)
		}
	})
	termui.Handle("/sys/kbd/<down>", func(termui.Event) {
		if slist.InSearchMode() {
			offset := slist.search.Next(slist.selectedIndex)
			slist.uilist.BorderLabel = slist.search.Title()
			slist.repaint(offset)
		} else {
			slist.repaint(1)
		}
	})
	termui.Handle("/sys/kbd/j", func(termui.Event) {
		if slist.InNormMode() {
			slist.repaint(1)
		} else {
			slist.appendQuery("j")
		}
	})
	termui.Handle("/sys/kbd/C-p", func(termui.Event) {
		if slist.InSearchMode() {
			offset := slist.search.Prev(slist.selectedIndex)
			slist.uilist.BorderLabel = slist.search.Title()
			slist.repaint(offset)
		} else {
			slist.repaint(-1)
		}
	})
	termui.Handle("/sys/kbd/<up>", func(termui.Event) {
		if slist.InSearchMode() {
			offset := slist.search.Prev(slist.selectedIndex)
			slist.uilist.BorderLabel = slist.search.Title()
			slist.repaint(offset)
		} else {
			slist.repaint(-1)
		}
	})
	termui.Handle("/sys/kbd/k", func(termui.Event) {
		if !slist.InSearchMode() {
			slist.repaint(-1)
		} else {
			slist.appendQuery("k")
		}
	})
	// vim page down
	termui.Handle("/sys/kbd/C-d", func(termui.Event) {
		if !slist.InSearchMode() {
			slist.repaint(10)
		}
	})
	// emacs page down
	termui.Handle("/sys/kbd/C-v", func(termui.Event) {
		if !slist.InSearchMode() {
			slist.repaint(10)
		}
	})
	// vim page up
	termui.Handle("/sys/kbd/C-u", func(termui.Event) {
		if !slist.InSearchMode() {
			slist.repaint(-10)
		}
	})
	// emacs page up
	termui.Handle("/sys/kbd/√", func(termui.Event) {
		if !slist.InSearchMode() {
			slist.repaint(-10)
		}
	})
	termui.Handle("/sys/kbd/G", func(termui.Event) {
		if !slist.InSearchMode() {
			//slist.repaint(-slist.selectedIndex - 1)
			slist.termWriter.Write([]byte("G"))
		} else {
			slist.appendQuery("G")
		}
	})
	termui.Handle("/sys/kbd/g", func(termui.Event) {
		if !slist.InSearchMode() {
			//			slist.repaint(-slist.selectedIndex)
			slist.termWriter.Write([]byte("g"))
		} else {
			slist.appendQuery("g")
		}
	})
	termui.Handle("/sys/wnd/resize", func(termui.Event) {
		slist.repaint(0)
	})
	termui.Handle("/sys/kbd", func(evt termui.Event) {
		kb, ok := evt.Data.(termui.EvtKbd)
		if !ok {
			return
		}
		if !slist.InSearchMode() {
			switch {
			// enter search mode
			case kb.KeyStr == "/" || kb.KeyStr == "C-s":
				slist.mode = ModeSearch
				slist.search.Reset()
				slist.uilist.BorderLabel = slist.search.Title()
				slist.repaint(0)
			case isNumber(kb.KeyStr):
				slist.termWriter.Write([]byte(kb.KeyStr))
			}

		} else {
			if kb.KeyStr == "C-8" {
				// delete char
				searchObj := slist.search
				_, size := utf8.DecodeLastRuneInString(searchObj.QueryStr)
				searchObj.QueryStr = searchObj.QueryStr[:len(searchObj.QueryStr)-size]
				slist.doSearch()
			} else if kb.KeyStr == "<space>" {
				slist.appendQuery(" ")
			} else {
				matched, _ := regexp.MatchString(`<.+>|C\-[^c]`, kb.KeyStr)
				if !matched {
					slist.appendQuery(kb.KeyStr)
				}
			}
		}
	})
}

// closeList release resources
func (sl *SelectList) closeList() {
	sl.termWriter.Stop()
}

func (slist *SelectList) writeTermLoop() {
	for term := range slist.termWriter.DataChan() {
		if slist.InNormMode() && term.IsMatched() {
			switch term.Name {
			case "0gg":
				slist.repaint(-slist.selectedIndex)
			case "gg":
				if idx, err := strconv.Atoi(strings.TrimSuffix(term.Text, "gg")); err == nil {
					slist.repaint(idx - 1 - slist.selectedIndex)
				}
			case "G":
				slist.repaint(-slist.selectedIndex - 1)

			}
		}
	}
}

type SearchObj struct {
	// filted results indices after searching
	SearchResultsIndices []int
	// cursor location, index of SearchResultsIndices
	SelectedResultIndex int
	// current query string
	QueryStr string
	// base title displayed when searching
	SearchTitle string
}

func (so *SearchObj) Reset() {
	so.SearchResultsIndices = []int{}
	so.SelectedResultIndex = 0
	so.QueryStr = ""
}

// render raw string as highlight format
func (sl *SelectList) Highlight(raw string, background bool) string {
	so := sl.search
	qs := strings.ToLower(so.QueryStr)
	raw1 := strings.ToLower(raw)
	if sl.InSearchMode() && so.QueryStr != "" && FuzzyContains(raw1, qs) {
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
	if len(so.SearchResultsIndices) > 0 {
		return fmt.Sprintf("%s%s     [Total %d items, current @%d, Navigation by C-n/C-p/down/up] press ESC exit search", so.SearchTitle, so.QueryStr, len(so.SearchResultsIndices), so.SelectedResultIndex+1)
	} else {
		return fmt.Sprintf("%s%s     Press ESC exit search", so.SearchTitle, so.QueryStr)
	}
}
func (so *SearchObj) Next(current int) int {
	if so.SelectedResultIndex >= 0 && so.SelectedResultIndex < len(so.SearchResultsIndices) {
		so.SelectedResultIndex = (so.SelectedResultIndex + 1) % len(so.SearchResultsIndices)
		return so.SearchResultsIndices[so.SelectedResultIndex] - current
	}
	return 0
}

func (so *SearchObj) Offset(current int) int {
	if so.SelectedResultIndex >= 0 && so.SelectedResultIndex < len(so.SearchResultsIndices) {
		return so.SearchResultsIndices[so.SelectedResultIndex] - current
	}
	return 0
}

func (so *SearchObj) Prev(current int) int {
	if so.SelectedResultIndex >= 0 && so.SelectedResultIndex < len(so.SearchResultsIndices) {
		so.SelectedResultIndex = (so.SelectedResultIndex - 1 + len(so.SearchResultsIndices)) % len(so.SearchResultsIndices)
		return so.SearchResultsIndices[so.SelectedResultIndex] - current
	}
	return 0
}

func (slist *SelectList) Show() {
	err := termui.Init()
	if err != nil {
		panic(err)
	}
	defer termui.Close()
	// set max line
	if slist.maxLine < termui.TermHeight()-3 {
		slist.maxLine = termui.TermHeight() - 3
	}

	slist.createUI()

	defer slist.closeList()

	go slist.writeTermLoop()

	slist.handleKeyboardEvents()

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

// format command for UI display
func (slist *SelectList) formatCommands() []string {
	matchedMap := make(map[int]int)
	searchObj := slist.search
	for i, j := range searchObj.SearchResultsIndices {
		matchedMap[j] = i
	}
	var strs []string
	start := slist.selectedIndex - slist.maxLine + 1
	if start < 0 {
		start = 0
	}
	end := start + slist.maxLine - 1
	if slist.InSearchMode() {
		start = searchObj.SelectedResultIndex - slist.maxLine + 1
		if start < 0 {
			start = 0
		}
		end = start + slist.maxLine - 1
	}
	var showIndex int
	for i, c := range slist.items {
		j, ok := matchedMap[i]
		if slist.InSearchMode() {
			if !ok || (j < start || j > end) {
				continue
			}
		} else if i < start || i > end {
			continue
		}
		if slist.InNormMode() || (slist.InSearchMode() && ok) {
			fmtI := "%02d"
			if len(slist.items) > 100 {
				fmtI = "%03d"
			}
			if i == slist.selectedIndex {
				strs = append(strs, fmt.Sprintf("["+fmtI+"] %s", showIndex+1, slist.Highlight(c, true)))
			} else {
				strs = append(strs, fmt.Sprintf("["+fmtI+"] %s", showIndex+1, slist.Highlight(c, false)))
			}
			showIndex++
		}
	}
	return strs
}
