package main

import (
	"os"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "relay"
	app.Usage = "command relay station"
	app.Authors = []cli.Author{
		{
			Name:  "JasonQu",
			Email: "qjpcpu@gmail.com",
		},
	}
	app.UsageText = "relay [global options] [command alias] [arguments...]"
	app.HideHelp = true
	app.HideVersion = true
	app.EnableBashCompletion = true
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "c",
			Usage: "specify config file",
			Value: os.Getenv("HOME") + "/.relay.conf",
		},
		cli.BoolFlag{
			Name:  "help, h",
			Usage: "show help",
		},
	}
	app.Action = func(c *cli.Context) error {
		if c.GlobalBool("help") {
			cli.ShowAppHelp(c)
			return nil
		}
		return runRelayCommand(newContext(c))
	}
	app.Commands = []cli.Command{
		{
			Name:  "!",
			Usage: "run last command",
			Action: func(c *cli.Context) error {
				return runLastCommand(newContext(c))
			},
		},
		{
			Name:  "@",
			Usage: "show relay history",
			Action: func(c *cli.Context) error {
				return runHistoryCommand(newContext(c))
			},
		},
	}
	app.Run(os.Args)
}
