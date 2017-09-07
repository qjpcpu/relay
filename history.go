package main

import (
	"encoding/json"
	"io/ioutil"
)

type Cache struct {
	LastIndex int
	Data      map[string]string
	History   []Cmd // name,realcommand
}

func loadCache() (c Cache, err error) {
	data, err := ioutil.ReadFile(cacheFile)
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &c)
	if err != nil {
		return
	}
	return
}

func saveCache(c Cache) {
	hmax := 50
	if l := len(c.History); l > hmax {
		c.History = c.History[(l - hmax):l]
	}
	data, _ := json.Marshal(c)
	ioutil.WriteFile(cacheFile, data, 0644)
}
