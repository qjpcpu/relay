package main

import (
	"os"

	"github.com/urfave/cli"
)

type context struct {
	*cli.Context
}

func newContext(c *cli.Context) *context {
	return &context{Context: c}
}
func (ctx *context) getConfigFile() string {
	if f := ctx.GlobalString("c"); isStrBlank(f) {
		return os.Getenv("HOME") + "/.relay.conf"
	} else {
		return f
	}
}

func (ctx *context) getCacheFile() string {
	return os.Getenv("HOME") + "/.relay_cache"
}

func (ctx *context) getAlias() string {
	return ctx.Args().Get(0)
}
