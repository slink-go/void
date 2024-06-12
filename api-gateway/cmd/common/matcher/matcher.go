package matcher

import (
	"regexp"
	"sort"
	"strings"
)

type PatternMatcher interface {
	Matches(input string) bool
	MatchesExact(input, exact string) bool
}

func NewPatternMatcher(patterns ...string) PatternMatcher {
	pm := patternMatcher{}
	if len(patterns) > 0 {
		sort.Slice(patterns, func(i, j int) bool {
			l1, l2 := len(patterns[i]), len(patterns[j])
			if l1 != l2 {
				return l1 < l2
			}
			return patterns[i] < patterns[j]
		})
	}
	var result []*regexp.Regexp
	for _, pattern := range patterns {
		result = append(result, regexp.MustCompile(strings.ReplaceAll(pattern, "*", ".*")))
	}
	pm.regularExpressions = result
	return &pm
}

type patternMatcher struct {
	regularExpressions []*regexp.Regexp
}

func (pm *patternMatcher) Matches(input string) bool {
	if pm.regularExpressions == nil || len(pm.regularExpressions) == 0 {
		return true
	}
	for _, regularExpression := range pm.regularExpressions {
		if regularExpression.MatchString(input) {
			return true
		}
	}
	return false
}
func (pm *patternMatcher) MatchesExact(input, exact string) bool {
	if strings.Contains(exact, "*") {
		return pm.Matches(input)
	} else {
		return input == exact
	}
}
