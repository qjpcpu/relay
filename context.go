package main

import (
	"os"
	"strings"

	"github.com/urfave/cli"
)

type context struct {
	*cli.Context
	isAlias bool
}

func newContext(c *cli.Context) *context {
	return &context{Context: c}
}
func (ctx *context) getConfigFile() string {
	if f := ctx.GlobalString("c"); IsBlankStr(f) {
		return os.Getenv("HOME") + "/.relay.conf"
	} else {
		return f
	}
}

func (ctx *context) getAlias() string {
	return ctx.Args().Get(0)
}

func (ctx *context) MarkAlias() {
	ctx.isAlias = true
}

func (ctx *context) ExtraArguments() []string {
	arguments := ctx.Args()
	if ctx.isAlias {
		if len(arguments) > 1 {
			return arguments[1:]
		}
	} else {
		return arguments
	}

	return nil
}

func IsBlankStr(s string) bool { return strings.TrimSpace(s) == "" }
