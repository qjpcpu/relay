package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/gizak/termui"
)

// ComplexSelect.Show panics on terminfo initialization errors. Retry with a
// compatible terminal only for the menu, preserving TERM for executed commands.
func showMenu(show func()) error {
	originalBuilder := termui.DefaultTxBuilder
	termui.DefaultTxBuilder = menuTextBuilder{originalBuilder}
	defer func() { termui.DefaultTxBuilder = originalBuilder }()

	err := tryShowMenu(show)
	if err == nil || err.Error() != "termbox: error while reading terminfo data: termbox: unsupported terminal" {
		return menuError(err)
	}

	term, wasSet := os.LookupEnv("TERM")
	fallback := "xterm-256color"
	if strings.HasPrefix(term, "tmux") {
		fallback = "screen-256color"
	}
	if err := os.Setenv("TERM", fallback); err != nil {
		return err
	}
	defer func() {
		if wasSet {
			os.Setenv("TERM", term)
		} else {
			os.Unsetenv("TERM")
		}
	}()
	return menuError(tryShowMenu(show))
}

// Keep the menu's green selection background, but avoid blue-on-green text
// whose contrast depends on the user's ANSI palette.
type menuTextBuilder struct {
	termui.TextBuilder
}

func (b menuTextBuilder) Build(s string, fg, bg termui.Attribute) []termui.Cell {
	cells := b.TextBuilder.Build(s, fg, bg)
	for i := range cells {
		if cells[i].Bg&0xFF != termui.ColorGreen {
			continue
		}
		switch cells[i].Fg & 0xFF {
		case termui.ColorBlue:
			cells[i].Fg = cells[i].Fg&^0xFF | termui.ColorBlack
		case termui.ColorWhite:
			// Search matches retain bold and get a distinct background.
			cells[i].Fg = cells[i].Fg&^0xFF | termui.ColorBlack
			cells[i].Bg = cells[i].Bg&^0xFF | termui.ColorWhite
		}
	}
	return cells
}

func menuError(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("cannot open selection menu (TERM=%q): %w", os.Getenv("TERM"), err)
}

func tryShowMenu(show func()) (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			initErr, ok := recovered.(error)
			if !ok || !strings.HasPrefix(initErr.Error(), "termbox: error while reading terminfo data:") {
				panic(recovered)
			}
			err = initErr
		}
	}()
	show()
	return nil
}
