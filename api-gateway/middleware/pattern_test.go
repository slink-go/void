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
			"test-1",
			"/api/service-a/test",
			"*/service-a/*",
			true,
		},
		{
			"test-2",
			"/api/service-a/test/path",
			"*/api/*",
			true,
		},
		{
			"test-3",
			"/service-a/api/test/path",
			"/service-a/api/*",
			true,
		},
		{
			"test-4",
			"/api/service-a/test",
			"*/service-a/*",
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
