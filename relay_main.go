package main

import (
	"fmt"
)

func runRelayCommand(c *context) (err error) {
	var commands []Cmd
	if commands, err = loadCommands(c); err != nil {
		return err
	}

	selectedCmd, ok := selectCommand(c, commands)
	if !ok {
		return nil
	}

	finalOptions, err := populateCommandWithCache(c, &selectedCmd)
	if err != nil {
		return err
	}

	// cache the comand as lastest command
	saveCache(c, selectedCmd, finalOptions)

	// run the command selected
	execCommand(c, selectedCmd.RealCommand)

	return nil
}

func selectCommand(ctx *context, commands []Cmd) (cmd Cmd, selected bool) {
	currentIndex, shortcut := findCommandByAlias(ctx, commands)

	menu := NewSelectListWithHints(currentIndex, commands2Items(commands), commands2Hints(commands))

	// if no shortcut specify, show selection UI
	if !shortcut {
		menu.Show()
	}
	if !menu.IsSelectNothing() {
		cmd = commands[menu.Selected()]
		selected = true
	}
	return
}

func populateCommandWithCache(ctx *context, cmd *Cmd) (finalOptions map[string]string, err error) {
	cache := shouldLoadCache(ctx)

	// populate command variables if exists
	if vlen := len(cmd.Variables()); vlen == 0 {
		fmt.Printf("\033[1;33m%s\033[0m\n\033[0;32m%s\033[0m\n", cmd.Name, cmd.RealCommand)
	} else {
		fmt.Printf("Fill command \033[1;33m%s\033[0m:\n", cmd.Name)
		optHistory := cache.GetOptionHistory(*cmd)
		if finalOptions, err = populateCommand(cmd, optHistory); err != nil {
			return nil, err
		}
		fmt.Printf("\033[0;32m%s\033[0m\n", cmd.RealCommand)
	}
	return
}

func runLastCommand(c *context) error {
	cache := shouldLoadCache(c)
	if len(cache.History) > 0 {
		cmd := cache.History[len(cache.History)-1]
		saveCache(c, cmd, nil)
		fmt.Printf("\033[1;33m%s\033[0m\n\033[0;32m%s\033[0m\n", cmd.Name, cmd.RealCommand)
		execCommand(c, cmd.RealCommand)
	} else {
		return runRelayCommand(c)
	}
	return nil
}

func runHistoryCommand(c *context) error {
	cache := shouldLoadCache(c)
	if len(cache.History) == 0 {
		fmt.Println("no history for now")
		return nil
	}
	history := make([]string, len(cache.History))
	historyNames := make([]string, len(cache.History))
	for i, c := range cache.History {
		history[len(cache.History)-i-1] = c.RealCommand
		historyNames[len(cache.History)-i-1] = c.Name + ": " + c.RealCommand
	}
	selects := NewSelectList(0, historyNames)
	selects.Show()
	if !selects.IsSelectNothing() {
		execCommand(c, history[selects.Selected()])
	}
	return nil
}
