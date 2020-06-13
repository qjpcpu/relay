package main

import (
	"fmt"
	"os"
	"strings"
)

func isStrBlank(s string) bool {
	return strings.TrimSpace(s) == ""
}

func fileDesc(file os.FileInfo) string {
	tp := PromptTypeFile
	if file.IsDir() {
		tp = PromptTypeDir
	}
	return fmt.Sprintf("%s mod:%s", tp, file.ModTime().Format("2006-01-02 15:04:05"))
}

func removeDupElem(list []string) []string {
	m := make(map[string]bool)
	var ret []string
	for _, v := range list {
		if !m[v] && !isStrBlank(v) {
			ret = append(ret, v)
		}
		m[v] = true
	}
	return ret
}
