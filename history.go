package main

import (
	"github.com/qjpcpu/common.v2/cli"
)

func saveCache(ctx *context, cmd Cmd) {
	db := cli.MustNewHomeFileDB("relay")
	defer db.Close()
	db.GetItemHistoryBucket("commands", 200).InsertItem(cmd)
}

func loadCache(ctx *context) (cmds []Cmd) {
	db := cli.MustNewHomeFileDB("relay")
	defer db.Close()
	db.GetItemHistoryBucket("commands", 200).ListItem(&cmds)
	return
}

func loadLatestCmd(ctx *context) *Cmd {
	if list := loadCache(ctx); len(list) > 0 {
		return &list[0]
	}
	return nil
}
