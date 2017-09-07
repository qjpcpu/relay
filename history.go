package main

import (
	"encoding/json"
	"io/ioutil"
)

type Cache struct {
	LastIndex int
	Data      map[string]string
	History   []string
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
	data, _ := json.Marshal(c)
	ioutil.WriteFile(cacheFile, data, 0644)
}
