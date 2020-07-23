package main

import (
	"os"

	"github.com/qjpcpu/common.v2/stringutil"
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
	if f := ctx.GlobalString("c"); stringutil.IsBlankStr(f) {
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
	if len(arguments) > 1 && ctx.isAlias {
		return arguments[1:]
	}
	return nil
}
