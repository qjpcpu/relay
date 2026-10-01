package main

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/gizak/termui"
)

func TestMenuTextContrast(t *testing.T) {
	builder := menuTextBuilder{termui.NewMarkdownTxBuilder()}
	cells := builder.Build("[选](fg-blue,bg-green,fg-underline)[中](fg-white,fg-bold,bg-green)[项](fg-cyan) [提示](fg-white)", termui.ColorCyan, termui.ColorDefault)
	if cells[0].Fg != termui.ColorBlack|termui.AttrUnderline || cells[0].Bg != termui.ColorGreen {
		t.Fatalf("selection colors: %+v", cells[0])
	}
	if cells[1].Fg != termui.ColorBlack|termui.AttrBold || cells[1].Bg != termui.ColorWhite {
		t.Fatalf("search match colors: %+v", cells[1])
	}
	if cells[2].Fg != termui.ColorCyan || cells[2].Bg != termui.ColorDefault || cells[4].Fg != termui.ColorWhite {
		t.Fatalf("other menu colors changed: %+v", cells)
	}
}

func TestShowMenuFallback(t *testing.T) {
	for _, tc := range []struct{ term, fallback string }{
		{"tmux-256color", "screen-256color"},
		{"unknown-terminal", "xterm-256color"},
	} {
		t.Run(tc.term, func(t *testing.T) {
			setTestTerm(t, tc.term)
			calls := 0
			err := showMenu(func() {
				calls++
				if calls == 1 {
					panic(errors.New("termbox: error while reading terminfo data: termbox: unsupported terminal"))
				}
				if term := os.Getenv("TERM"); term != tc.fallback {
					t.Fatalf("fallback TERM = %q, want %q", term, tc.fallback)
				}
			})
			if err != nil || calls != 2 || os.Getenv("TERM") != tc.term {
				t.Fatalf("err=%v calls=%d restored TERM=%q", err, calls, os.Getenv("TERM"))
			}
		})
	}
}

func TestShowMenuNormal(t *testing.T) {
	setTestTerm(t, "xterm-256color")
	calls := 0
	if err := showMenu(func() { calls++ }); err != nil || calls != 1 {
		t.Fatalf("err=%v calls=%d", err, calls)
	}
}

func TestShowMenuFailedRetry(t *testing.T) {
	setTestTerm(t, "tmux-256color")
	calls := 0
	err := showMenu(func() {
		calls++
		panic(errors.New("termbox: error while reading terminfo data: termbox: unsupported terminal"))
	})
	if err == nil || calls != 2 || os.Getenv("TERM") != "tmux-256color" {
		t.Fatalf("err=%v calls=%d restored TERM=%q", err, calls, os.Getenv("TERM"))
	}
}

func TestShowMenuOtherTerminfoError(t *testing.T) {
	calls := 0
	err := showMenu(func() {
		calls++
		panic(errors.New("termbox: error while reading terminfo data: EOF"))
	})
	if err == nil || !strings.Contains(err.Error(), "EOF") || calls != 1 {
		t.Fatalf("err=%v calls=%d", err, calls)
	}
}

func TestShowMenuUnrelatedPanic(t *testing.T) {
	defer func() {
		if got := recover(); got != "unrelated bug" {
			t.Fatalf("panic = %v", got)
		}
	}()
	showMenu(func() { panic("unrelated bug") })
}

func setTestTerm(t *testing.T, term string) {
	t.Helper()
	original, wasSet := os.LookupEnv("TERM")
	if err := os.Setenv("TERM", term); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if wasSet {
			os.Setenv("TERM", original)
		} else {
			os.Unsetenv("TERM")
		}
	})
}
