package main

import (
	"bytes"
	"strings"
	"unsafe"

	gpy "github.com/mozillazg/go-pinyin"
)

func Contain(text string, keyword string) (substr string, index int) {
	s := textToSentence(Args{StrictMatch: true}, text)
	return s.find(keywordToBytes(keyword))
}

func FuzzyContain(text string, keyword string) (substr string, index int) {
	s := textToSentence(Args{}, text)
	return s.find(keywordToBytes(keyword))
}

type word []byte
type term struct {
	offset    int
	raw       []byte
	isChinese bool
	alias     []word
}
type sentence struct {
	rawText string
	terms   []*term
	args    *Args
}

func (s *sentence) getRaw(t *term) string {
	return s.rawText[t.offset : t.offset+len(t.raw)]
}

func (s *sentence) find(keyword []byte) (string, int) {
	if len(keyword) == 0 || len(s.terms) == 0 {
		return "", -1
	}
	for i := 0; i < len(s.terms); i++ {
		if termIdx := findKeyword(*s.args, s.terms, i, keyword); termIdx != -1 {
			var ret string
			for j := i; j <= termIdx; j++ {
				ret += s.getRaw(s.terms[j])
				if nextOffset := s.terms[j].offset + len(s.terms[j].raw); j < termIdx && nextOffset < s.terms[j+1].offset {
					ret += s.rawText[nextOffset:s.terms[j+1].offset]
				}
			}
			return ret, s.terms[i].offset
		}
	}
	return "", -1
}

type Args struct {
	StrictMatch bool
}

func findKeyword(args Args, terms []*term, termIdx int, keyword []byte) int {
	term := terms[termIdx]
	for _, alias := range term.alias {
		size := matchBytes(alias, keyword)
		if size == 0 {
			continue
		} else if size < len(keyword) && size < len(alias) {
			continue
		} else if size == len(keyword) {
			if size < len(alias) {
				if !args.StrictMatch {
					return termIdx
				}
			} else {
				return termIdx
			}
		} else if size == len(alias) {
			if size < len(keyword) {
				if termIdx+1 >= len(terms) {
					continue
				}
				if idx := findKeyword(args, terms, termIdx+1, keyword[size:]); idx != -1 {
					return idx
				}
			} else {
				return termIdx
			}
		}
	}
	return -1
}

func keywordToBytes(keyword string) (data []byte) {
	pargs := gpy.NewArgs()
	runes := []rune(keyword)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		py := gpy.SinglePinyin(r, pargs)
		if len(py) > 0 {
			data = append(data, ([]byte(py[0]))...)
		} else {
			if isCharOrNumber(r) {
				data = append(data, runesToBytes([]rune{r})...)
			}
		}
	}
	return
}

func textToSentence(args Args, text string) *sentence {
	var terms []*term
	pargs := gpy.NewArgs()
	pargs.Heteronym = true

	runes := []rune(text)
	for i := 0; i < len(runes); {
		r := runes[i]
		py := gpy.SinglePinyin(r, pargs)
		if len(py) > 0 {
			t := makePinyinTerm(r, py, args.StrictMatch)
			t.offset = len(runesToBytes(runes[:i]))
			terms = append(terms, t)
			i++
		} else {
			if isCharOrNumber(r) {
				j := i + 1
				for ; j < len(runes) && isCharOrNumber(runes[j]); j++ {
				}
				data := []byte(strings.ToLower(string(runes[i:j])))
				t := &term{
					offset:    len(runesToBytes(runes[:i])),
					raw:       data,
					isChinese: false,
					alias:     []word{data},
				}
				terms = append(terms, t)
				i = j
			} else {
				i++
			}
		}
	}
	sen := &sentence{rawText: text, terms: terms, args: &args}
	return sen
}

func isCharOrNumber(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
}

func runesToBytes(r []rune) []byte {
	str := string(r)
	return *((*[]byte)((unsafe.Pointer(&str))))
}

// 声母表
var initialArray = strings.Split(
	"b,p,m,f,d,t,n,l,g,k,h,j,q,x,r,zh,ch,sh,z,c,s",
	",",
)

func makePinyinTerm(r rune, pinyins []string, isStrict bool) *term {
	var words []word
	for _, w := range pinyins {
		data := []byte(strings.ToLower(w))
		words = append(words, data)
		if !isStrict {
			for _, c := range initialArray {
				cc := []byte(c)
				if bytes.HasPrefix(data, cc) {
					words = append(words, cc)
				}
			}
		}
	}
	return &term{
		raw:       runesToBytes([]rune{r}),
		isChinese: true,
		alias:     words,
	}
}

func matchBytes(left, right []byte) (length int) {
	for i := 0; i < len(left) && i < len(right) && left[i] == right[i]; i++ {
		length++
	}
	return
}
