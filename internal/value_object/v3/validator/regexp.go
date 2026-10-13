package validator

import (
	"regexp"
	"sync"
)

const (
	splitParamsRegexString = `'[^']*'|\S+`
)

func lazyRegexCompile(str string) func() *regexp.Regexp {
	var regex *regexp.Regexp
	var once sync.Once
	return func() *regexp.Regexp {
		once.Do(func() {
			regex = regexp.MustCompile(str)
		})
		return regex
	}
}

var splitParamsRegex = lazyRegexCompile(splitParamsRegexString)
