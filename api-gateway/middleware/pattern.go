package middleware

import (
	"regexp"
	"strings"
)

func Match(source, pattern string, re *regexp.Regexp) bool {
	if strings.Contains(pattern, "*") {
		return re.Match([]byte(source))
	} else {
		return source == pattern
	}
}
