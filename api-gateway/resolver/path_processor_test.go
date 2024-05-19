package resolver

import (
	"errors"
	"testing"
)

func TestPartsIsEmpty(t *testing.T) {
	tests := []struct {
		name           string
		input          []string
		expectedResult bool
	}{
		{
			"nil parts array test",
			nil,
			true,
		},
		{
			"empty parts array test",
			[]string{},
			true,
		},
		{
			"parts array empty values test",
			[]string{"", ""},
			true,
		},
		{
			"non-empty parts array test",
			[]string{"part1", "part2"},
			false,
		},
	}
	pp := &pathProcessor{}
	for _, tt := range tests {
		res := pp.partsIsEmpty(tt.input)
		if res != tt.expectedResult {
			t.Fatalf("[%s] expected '%v' got '%v'", tt.name, tt.expectedResult, res)
		}
	}
}

func TestPartsSplit(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedResult []string
		expectedError  error
	}{
		{
			"empty path test",
			"",
			nil,
			NewErrInvalidPath(""),
		},
		{
			"invalid path test 1",
			"///",
			nil,
			NewErrInvalidPath(""),
		},
		{
			"invalid path test 2",
			"/",
			nil,
			NewErrInvalidPath(""),
		},
		{
			"invalid path test 3",
			"/api",
			nil,
			NewErrInvalidPath(""),
		},
		{
			"invalid path test 3",
			"/api/",
			nil,
			NewErrInvalidPath(""),
		},
		{
			"valid path test 1",
			"/api/service",
			[]string{"service", "api"},
			nil,
		},
		{
			"valid path test 2",
			"/service/api",
			[]string{"service", "api"},
			nil,
		},
		{
			"valid path test 3",
			"/api/service/api",
			[]string{"service", "api"},
			nil,
		},
		{
			"valid path test 4",
			"/api/service/test",
			[]string{"service", "api", "test"},
			nil,
		},
		{
			"valid path test 5",
			"/service/api/test",
			[]string{"service", "api", "test"},
			nil,
		},
		{
			"valid path test 6",
			"/api/service/api/test",
			[]string{"service", "api", "test"},
			nil,
		},
	}
	pp := &pathProcessor{}
	for _, tt := range tests {
		testPartsSplit(t, pp, tt.name, tt.input, tt.expectedResult, tt.expectedError)
	}
}
func testPartsSplit(t *testing.T, pp PathProcessor, test, input string, expectedResult []string, expectedError error) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("[%s] paniced: %v]", test, r)
		}
	}()
	parts, err := pp.Split(input)
	if expectedError != nil {
		checkExpectedError(t, test, expectedError, err)
	} else {
		checkTestArrayResult(t, test, expectedResult, parts, err)
	}
}

func TestPartsJoin(t *testing.T) {
	tests := []struct {
		name           string
		inputUrl       string
		inputParts     []string
		expectedResult string
		expectedError  error
	}{
		{
			"empty input test",
			"",
			nil,
			"",
			NewErrEmptyBaseUrl(),
		},
		{
			"empty parts test",
			"localhost:1234",
			nil,
			"localhost:1234",
			nil,
		},
		{
			"empty url test",
			"",
			[]string{"a", "b", "c"},
			"",
			NewErrEmptyBaseUrl(),
		},
		{
			"valid input test",
			"localhost:1234",
			[]string{"a", "b", "c"},
			"localhost:1234/b/c",
			nil,
		},
	}
	pp := &pathProcessor{}
	for _, tt := range tests {
		testPartsJoin(t, pp, tt.name, tt.inputUrl, tt.inputParts, tt.expectedResult, tt.expectedError)
	}
}
func testPartsJoin(t *testing.T, pp PathProcessor, test, url string, parts []string, expectedResult string, expectedError error) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("[%s] paniced: %v]", test, r)
		}
	}()
	result, err := pp.Join(url, parts)
	if expectedError != nil {
		checkExpectedError(t, test, expectedError, err)
	} else {
		checkTestResult(t, test, expectedResult, result, err)
	}
}

func TestPathResolve(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedResult string
		expectedError  error
	}{
		{
			"empty input test",
			"",
			"",
			NewErrInvalidPath(""),
		},
		{
			"invalid input test 1",
			"/api",
			"",
			NewErrInvalidPath(""),
		},
		{
			"invalid input test 2",
			"/api/",
			"",
			NewErrInvalidPath(""),
		},
		{
			"invalid input test 3",
			"/api/service-c/test",
			"",
			NewErrServiceUnavailable(""),
		},
		{
			"valid input test 1",
			"/api/a",
			"http://service-a/api",
			nil,
		},
		{
			"valid input test 2",
			"/a/api",
			"http://service-a/api",
			nil,
		},
		{
			"valid input test 3",
			"/a/api/b",
			"http://service-a/api/b",
			nil,
		},
		{
			"valid input test 4",
			"/api/a/api/b",
			"http://service-a/api/b",
			nil,
		},
	}

	var mappings = make(map[string][]string)
	mappings["a"] = []string{"service-a"}
	mappings["b"] = []string{"service-b"}
	pp := &pathProcessor{}
	for _, tt := range tests {
		testPathResolve(t, pp, NewServiceResolver(NewStaticServiceRegistry(mappings)), tt.name, tt.input, tt.expectedResult, tt.expectedError)
	}
}
func testPathResolve(t *testing.T, pp PathProcessor, resolver ServiceResolver, test string, input string, expectedResult string, expectedError error) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("[%s] paniced: %v]", test, r)
		}
	}()
	result, err := pp.UrlResolve(input, resolver)
	if expectedError != nil {
		checkExpectedError(t, test, expectedError, err)
	} else {
		checkTestResult(t, test, expectedResult, result, err)
	}
}

func checkExpectedError(t *testing.T, test string, expected, actual error) {
	if actual == nil {
		t.Errorf("[%s] expected error, but not happened", test)
	} else if !errors.Is(actual, expected) {
		t.Errorf("[%s] expected error '%T', got '%T'", test, expected, actual)
	}
}
func checkTestResult(t *testing.T, test string, expected string, actual string, err error) {
	if err != nil {
		t.Fatalf("[%s] unexpected error: %s", test, err)
	}
	if actual != expected {
		t.Fatalf("[%s] expected '%v' got '%v'", test, expected, actual)
	}
}
func checkTestArrayResult(t *testing.T, test string, expected, actual []string, err error) {
	if err != nil {
		t.Fatalf("[%s] unexpected error: %s", test, err)
	}
	if !arraysMatch(actual, expected) {
		t.Fatalf("[%s] expected '%v' got '%v'", test, expected, actual)
	}
}

func arraysMatch(a, b []string) bool {
	if isEmpty(a) && isEmpty(b) {
		return true
	}
	if isEmpty(a) && !isEmpty(b) || !isEmpty(a) && isEmpty(b) {
		return false
	}
	if len(a) != len(b) {
		return false
	}
	matches := 0
	for i, _ := range a {
		if a[i] == b[i] {
			matches++
		}
	}
	return matches == len(a)
}
func isEmpty(array []string) bool {
	return array == nil || len(array) == 0
}
