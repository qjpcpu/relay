package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
)

type Cache struct {
	History []Cmd
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
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	encoder.Encode(c)
	ioutil.WriteFile(cacheFile, buffer.Bytes(), 0644)
}
