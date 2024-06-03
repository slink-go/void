package security

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/slink-go/api-gateway/middleware/constants"
	"github.com/slink-go/api-gateway/resolver"
	"github.com/slink-go/logging"
	"io"
	"net/http"
)

type tokenUserDetailsProvider struct {
	authProvider    AuthProvider
	serviceResolver resolver.ServiceResolver
	pathProcessor   resolver.PathProcessor
	endpoint        string
	method          string
	logger          logging.Logger
	responseParser  ResponseParser
}

func NewTokenBasedUserDetailsProvider() UserDetailsProvider {
	udp := tokenUserDetailsProvider{
		logger: logging.GetLogger("rest-user-details-provider"),
		method: "GET",
	}
	return &udp
}
func (p *tokenUserDetailsProvider) WithAuthProvider(provider AuthProvider) UserDetailsProvider {
	p.authProvider = provider
	return p
}
func (p *tokenUserDetailsProvider) WithServiceResolver(resolver resolver.ServiceResolver) UserDetailsProvider {
	p.serviceResolver = resolver
	return p
}
func (p *tokenUserDetailsProvider) WithPathProcessor(processor resolver.PathProcessor) UserDetailsProvider {
	p.pathProcessor = processor
	return p
}
func (p *tokenUserDetailsProvider) WithAuthEndpoint(endpoint string) UserDetailsProvider {
	p.endpoint = endpoint
	return p
}
func (p *tokenUserDetailsProvider) WithMethod(endpoint string) UserDetailsProvider {
	p.endpoint = endpoint
	return p
}
func (p *tokenUserDetailsProvider) WithResponseParser(responseParser ResponseParser) UserDetailsProvider {
	p.responseParser = responseParser
	return p
}

func (p *tokenUserDetailsProvider) Get(token string) (UserDetails, error) {

	if token == "" {
		return nil, errors.New("auth token is not provided")
	}
	if p.authProvider == nil {
		return nil, errors.New("auth provider is not set")
	}
	if p.serviceResolver == nil {
		return nil, errors.New("service resolver is not set")
	}
	if p.pathProcessor == nil {
		return nil, errors.New("path processor is not set")
	}
	if p.endpoint == "" {
		return nil, errors.New("auth endpoint is not set")
	}

	service, err := p.pathProcessor.HostResolve(p.endpoint, p.serviceResolver)
	if err != nil {
		p.logger.Debug("could not resolve auth endpoint: %s", err)
		p.logger.Debug("will use raw auth endpoint: %s", p.endpoint)
		service = p.endpoint
	} else {
		p.logger.Debug("resolved auth endpoint: %s", service)
	}

	res, err := p.exchange(service, fmt.Sprintf("Bearer %s", token))
	if err != nil {
		return nil, err
	}
	return p.processAuthResponse(res)

}

func (p *tokenUserDetailsProvider) exchange(endpoint, header string) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest(p.method, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("could not generate delegated authorization request: %s", err)
	}
	req.Header.Set(constants.HdrAuthorization, header)
	return client.Do(req)
}
func (p *tokenUserDetailsProvider) processAuthResponse(res *http.Response) (UserDetails, error) {
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode > 399 {
		return nil, fmt.Errorf("%v", body)
	}
	authData := make(map[string]interface{})
	if err = json.Unmarshal(body, &authData); err != nil {
		return nil, err
	}
	result := p.responseParser.Parse(authData)
	return result, nil
}
