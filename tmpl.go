package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/hoisie/mustache"
)

func (cmd *Cmd) Variables() []string {
	re := regexp.MustCompile(`{{\s*([a-zA-Z_0-9]+)\s*}}`)
	matchers := re.FindAllString(cmd.Cmd, -1)
	var vars []string
	for _, m := range matchers {
		v := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(m, `{{`), `}}`))
		if v != "" {
			vars = append(vars, v)
		}
	}
	return vars
}

func (cmd *Cmd) Populate(data map[string]string) {
	tmpl, err := mustache.ParseString(cmd.Cmd)
	if err != nil {
		return
	}
	for _, vn := range cmd.Variables() {
		v1, ok1 := data[vn]
		if v2, ok2 := cmd.Defaults[vn]; ok2 && v2 != "" && (!ok1 || v1 == "") {
			data[vn] = v2
		}
	}
	cmd.RealCommand = tmpl.Render(data)
}

func populateCommand(cmd *Cmd, datas ...map[string]string) map[string]string {
	variables := cmd.Variables()
	if len(variables) == 0 {
		return nil
	}
	if len(datas) == 1 && len(datas[0]) > 0 {
		cmd.Populate(datas[0])
		return datas[0]
	}
	data := make(map[string]string)
	reader := bufio.NewReader(os.Stdin)
	for _, v := range variables {
		if _, ok := data[v]; ok {
			continue
		}
		if defaultVal := cmd.Defaults[v]; defaultVal != "" {
			fmt.Printf("%s(%s): ", v, defaultVal)
		} else {
			fmt.Printf("%s: ", v)
		}
		text, _ := reader.ReadString('\n')
		text = strings.Replace(text, "\n", "", -1)
		data[v] = strings.TrimSpace(text)
	}
	cmd.Populate(data)
	return data
}
