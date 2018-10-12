package main

import (
	"github.com/mozillazg/go-pinyin"
	"strings"
)

func FuzzyContains(s string, substr string) bool {
	if strings.Contains(s, substr) {
		return true
	}
	idx, _ := PinyinContains(s, substr)
	return idx >= 0
}

func FuzzyIndex(s string, substr string) (int, string) {
	if idx := strings.Index(s, substr); idx >= 0 {
		return idx, substr
	}
	return PinyinContains(s, substr)
}

func ToPinyin(raw string) (string, [][2]int) {
	args := pinyin.NewArgs()
	hanzi := make([][2]int, 0)
	var text string
	var idx int
	for _, r := range raw {
		py := pinyin.SinglePinyin(r, args)
		if len(py) > 0 {
			text += py[0]
			hanzi = append(hanzi, [2]int{idx, len(py[0])})
			idx += len(py[0])
		} else {
			text += string(r)
			idx++
		}
	}
	return text, hanzi
}

func PinyinContains(raw string, term string) (int, string) {
	py, words := ToPinyin(raw)
	term, _ = ToPinyin(term)
	if !strings.Contains(py, term) {
		return -1, ""
	}
	for {
		idx := strings.Index(py, term)
		if idx < 0 {
			break
		}
		valid := true
		for _, w := range words {
			if idx > w[0] && idx < w[0]+w[1]-1 {
				r := []rune(py)
				r[idx] = '\r'
				py = string(r)
				valid = false
				break
			}
		}
		if valid {
			end := idx + len(term)
			for _, w := range words {
				if end-1 >= w[0] && end <= w[0]+w[1] {
					end = w[0] + w[1]
					break
				}
			}
			start_offset, end_offset := 0, 0
			for _, w := range words {
				if w[0] > end {
					break
				}
				if idx > w[0] {
					start_offset = start_offset + w[1] - 1
					end_offset = end_offset + w[1] - 1
				} else if w[0]+w[1] <= end {
					end_offset = end_offset + w[1] - 1
				}
			}
			b := []rune(raw)
			matched := string(b[idx-start_offset : end-end_offset])
			return strings.Index(raw, matched), matched
		}
	}
	return -1, ""
}
