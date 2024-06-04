package middleware

import "testing"

func TestPatternMatch(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		pattern string
		result  bool
	}{
		{
			"test 1",
			"http://host/api/service-a/test",
			"*/service-a/*",
			true,
		},
		{
			"test 2",
			"http://host/api/service-a/test",
			"*/api/*",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Match(tt.input, tt.pattern)
			if result != tt.result {
				t.Fail()
			}
		})
	}
}
