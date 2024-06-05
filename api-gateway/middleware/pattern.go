package middleware

import (
	"regexp"
	"strings"
)

func Match(source, pattern string) bool {
	// TODO: implement URL pattern matching check
	if strings.Contains(pattern, "*") {
		pattern = strings.ReplaceAll(pattern, "*", ".*")
		re := regexp.MustCompile(pattern)
		return re.Match([]byte(source))
	} else {
		return source == pattern
	}
}
