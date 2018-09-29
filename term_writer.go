package main

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
)

type Term struct {
	Name         string
	PrefixRegexp string
	SuffixSymbol string
}

type MatchedTerm struct {
	Text string
	Term
}

func (mt MatchedTerm) IsMatched() bool {
	return mt != MatchedTerm{}
}

type compiledTerm struct {
	fullPrefix  *regexp.Regexp
	startPrefix *regexp.Regexp
	Term
}

type compiledTerms []compiledTerm

func (ct compiledTerms) HasTerm(name string) (int, bool) {
	for i, t := range ct {
		if t.Name == name {
			return i, true
		}
	}
	return -1, false
}

func (ct compiledTerm) MatchPrefix(termText string) bool {
	if ct.Term.PrefixRegexp == "" {
		return true
	}
	return ct.fullPrefix.MatchString(termText)
}

func (ct compiledTerm) IsPrefixOf(termText string) bool {
	if ct.Term.PrefixRegexp == "" {
		return true
	}
	return ct.startPrefix.MatchString(termText)
}

func (ct compiledTerms) MaybeMatch(termText string) bool {
	for _, term := range ct {
		length := len(termText)
		if term.Term.PrefixRegexp == "" {
			if strings.HasSuffix(term.SuffixSymbol, termText) && termText != "" {
				return true
			} else {
				continue
			}
		}
		if !term.IsPrefixOf(termText) {
			continue
		}
		for i := 0; i <= length; i++ {
			prefix, suffix := termText[:length-i], termText[length-i:]
			if len(suffix) <= len(term.SuffixSymbol) && term.MatchPrefix(prefix) && strings.HasPrefix(term.SuffixSymbol, suffix) {
				return true
			}
		}
	}
	return false
}

func (ct compiledTerms) Match(termText string) (MatchedTerm, bool) {
	for _, term := range ct {
		if term.Term.PrefixRegexp == "" {
			if termText == term.SuffixSymbol {
				return MatchedTerm{
					Text: termText,
					Term: term.Term,
				}, true
			} else {
				continue
			}
		}
		prefix := strings.TrimSuffix(termText, term.SuffixSymbol)
		if term.MatchPrefix(prefix) && strings.HasSuffix(termText, term.SuffixSymbol) {
			return MatchedTerm{
				Text: termText,
				Term: term.Term,
			}, true
		}
	}
	return MatchedTerm{}, false
}

type TermWriter struct {
	lastTypeAt time.Time
	data       bytes.Buffer
	terms      compiledTerms
	resChan    chan MatchedTerm
	stop       bool
	mutex      *sync.Mutex
}

func NewTermWriter() *TermWriter {
	b := &TermWriter{
		lastTypeAt: time.Now(),
		resChan:    make(chan MatchedTerm, 1),
		mutex:      new(sync.Mutex),
	}
	return b
}

func (b *TermWriter) AddTerm(name string, prefixReg string, endingText string) error {
	if _, ok := b.terms.HasTerm(name); ok {
		return fmt.Errorf("already contains term:%s", name)
	}
	if endingText == "" {
		return fmt.Errorf("endingText should not be empty")
	}
	modifyReg := prefixReg
	if !strings.HasSuffix(modifyReg, "$") {
		modifyReg += "$"
	}
	if !strings.HasPrefix(modifyReg, "^") {
		modifyReg = "^" + modifyReg
	}
	expr, err := regexp.Compile(modifyReg)
	if err != nil {
		return err
	}
	expr1, _ := regexp.Compile(strings.TrimSuffix(modifyReg, "$"))
	b.terms = append(b.terms, compiledTerm{
		fullPrefix:  expr,
		startPrefix: expr1,
		Term: Term{
			Name:         name,
			PrefixRegexp: prefixReg,
			SuffixSymbol: endingText,
		},
	})
	return nil
}

func (b *TermWriter) DataChan() <-chan MatchedTerm {
	return b.resChan
}

func (b *TermWriter) Stop() {
	if b.stop {
		return
	}
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.stop = true
	close(b.resChan)
}

func (b *TermWriter) Write(data []byte) (int, error) {
	if b.stop {
		return 0, fmt.Errorf("stopped")
	}
	b.mutex.Lock()
	defer b.mutex.Unlock()
	now := time.Now()
	if now.Sub(b.lastTypeAt) > 2*time.Second {
		b.data.Reset()
	}
	b.lastTypeAt = now
	n, err := b.data.Write(data)
	text := b.data.String()
	if term, matched := b.terms.Match(text); matched {
		b.data.Reset()
		b.resChan <- term
	}
	if !b.terms.MaybeMatch(text) {
		b.data.Reset()
	}
	return n, err
}
