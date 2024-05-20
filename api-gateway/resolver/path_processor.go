package resolver

import (
	"fmt"
	"github.com/slink-go/api-gateway/registry"
	"strings"
)

type PathProcessor interface {
	Split(input string) ([]string, error)
	Join(serviceUrl string, parts []string) (string, error)
	UrlResolve(input string, resolver ServiceResolver) (string, error)
}

func NewPathProcessor() PathProcessor {
	return &pathProcessor{}
}

type pathProcessor struct {
}

func (pp *pathProcessor) Split(input string) ([]string, error) {

	parts := strings.Split(
		strings.TrimSuffix(
			strings.TrimPrefix(
				strings.ToLower(input),
				"/",
			),
			"/",
		),
		"/",
	)
	if pp.partsIsEmpty(parts) {
		return nil, NewErrInvalidPath(input)
	}

	useApi := false
	if parts[0] == "api" {
		useApi = true
		parts = parts[1:]
	}

	if len(parts) == 0 {
		return nil, NewErrInvalidPath(input)
	}

	if useApi {
		if len(parts) == 1 {
			parts = append(parts, "api")
		} else if parts[1] != "api" {
			parts = append(parts[:2], parts[1:]...)
			parts[1] = "api"
		}
	}

	return parts, nil
}
func (pp *pathProcessor) Join(serviceUrl string, parts []string) (string, error) {
	if serviceUrl == "" {
		return "", NewErrEmptyBaseUrl()
	}
	if len(parts) == 0 {
		return serviceUrl, nil
	}
	return fmt.Sprintf("%s/%s", serviceUrl, strings.Join(parts[1:], "/")), nil
}
func (pp *pathProcessor) UrlResolve(input string, resolver ServiceResolver) (string, error) {
	parts, err := pp.Split(input)
	if err != nil {
		return "", err
	}

	target, err := resolver.Resolve(parts[0])
	if err != nil {
		return "", err
	}
	if target == "" {
		return "", registry.NewErrServiceUnavailable(parts[0])
	}

	url, err := pp.Join(target, parts)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("http://%s", url), nil
}

func (pp *pathProcessor) partsIsEmpty(parts []string) bool {
	for _, part := range parts {
		if part != "" {
			return false
		}
	}
	return true
}
