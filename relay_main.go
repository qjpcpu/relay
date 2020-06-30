package main

import (
	"fmt"

	"github.com/qjpcpu/common.v2/cli"
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

	if err = populateCommandWithCache(c, &selectedCmd); err != nil {
		return err
	}

	// cache the comand as lastest command
	saveCache(c, selectedCmd)

	confirmComand(c, &selectedCmd)

	// run the command selected
	execCommand(c, selectedCmd.RealCommand)

	return nil
}

func selectCommand(ctx *context, commands []Cmd) (cmd Cmd, selected bool) {
	currentIndex, shortcut := findCommandByAlias(ctx, commands)

	menu := cli.NewComplexSelectWithHints(currentIndex, commands2Items(commands), commands2Hints(commands))

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

func populateCommandWithCache(ctx *context, cmd *Cmd) (err error) {
	// populate command variables if exists
	if vlen := len(cmd.Variables()); vlen == 0 {
		fmt.Printf("\033[1;33m%s\033[0m\n\033[0;32m%s\033[0m\n", cmd.Name, cmd.RealCommand)
	} else {
		fmt.Printf("Fill command \033[1;33m%s\033[0m:\n", cmd.Name)
		if err = populateCommand(cmd); err != nil {
			return err
		}
		fmt.Printf("\033[0;32m%s\033[0m\n", cmd.RealCommand)
	}
	return
}

func runLastCommand(c *context) error {
	if cmd := loadLatestCmd(c); cmd != nil {
		saveCache(c, *cmd)
		fmt.Printf("\033[1;33m%s\033[0m\n\033[0;32m%s\033[0m\n", cmd.Name, cmd.RealCommand)
		execCommand(c, cmd.RealCommand)
	} else {
		return runRelayCommand(c)
	}
	return nil
}

func runHistoryCommand(c *context) error {
	commandList := loadCache(c)
	if len(commandList) == 0 {
		fmt.Println("no history for now")
		return nil
	}
	history := make([]string, len(commandList))
	historyNames := make([]string, len(commandList))
	for i, c := range commandList {
		history[i] = c.RealCommand
		historyNames[i] = c.Name + ": " + c.RealCommand
	}
	selects := cli.NewComplexSelect(0, historyNames)
	selects.Show()
	if !selects.IsSelectNothing() {
		execCommand(c, history[selects.Selected()])
	}
	return nil
}
