package main

import (
	"testing"
)

func TestPinyin(t *testing.T) {
	checkContains := func(s, term, matched string) {
		idx, size := PinyinContains(s, term)
		if matched != "" {
			if idx < 0 {
				t.Fatalf("[%s] should contains [%s] with %s,but failed", s, term, matched)
				return
			}
			if s[idx:idx+size] != matched {
				t.Fatalf("[%s] should contains [%s] with %s, but get %s", s, term, matched, s[idx:idx+size])
				return
			}
		} else {
			if idx >= 0 {
				t.Fatalf("[%s] should not contains [%s], but get %s", s, term, s[idx:idx+size])
			}
		}
	}
	checkContains(`测试环境`, "i", "")
	checkContains(`测试环境`, "shi", "试")
	checkContains(`测试环境`, "shi环", "试环")
	checkContains(`测试环境`, "测sh", "测试")
	checkContains("ip地址", "pd", "p地")
}
