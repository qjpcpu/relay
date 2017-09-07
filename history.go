package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
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
