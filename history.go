package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"strings"
)

type Cache struct {
	LastIndex int
	Data      map[string]string
}

func loadCache() (c Cache, err error) {
	data, err := ioutil.ReadFile(cacheFile)
	if err != nil {
		return
	}
	err = yaml.Unmarshal(data, &c)
	if err != nil {
		return
	}
	return
}

func saveCache(c Cache) {
	data, _ := yaml.Marshal(c)
	ioutil.WriteFile(cacheFile, data, 0644)
}

func changeCliHistory(prefix string, newcmd string) {
	file := os.Getenv("HISTFILE")
	if file == "" {
		home := os.Getenv("HOME")
		shell := strings.Split(os.Getenv("SHELL"), "/")
		file = home + "/." + shell[len(shell)-1] + "_history"
	}
	if file == "" {
		return
	}
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return
	}
	lines := strings.Split(string(data), "\n")
	var cmd string
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.Contains(lines[i], prefix) {
			cmd = lines[i]
			index := strings.Index(cmd, prefix)
			if index < 0 {
				return
			}
			lines[i] = cmd[0:index] + newcmd
			break
		}
	}
	if cmd == "" {
		return
	}
	str := strings.Join(lines, "\n")
	ioutil.WriteFile(file, []byte(str), 0600)
}
