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

// region - option

type UserDetailsProviderOption interface {
	apply(p *tokenBasedUserDetailsProvider)
}

// region -> auth provider

type authProviderOption struct {
	value AuthProvider
}

func (o *authProviderOption) apply(p *tokenBasedUserDetailsProvider) {
	if o.value != nil {
		p.authProvider = o.value
	}
}

func UdpWithAuthProvider(value AuthProvider) UserDetailsProviderOption {
	return &authProviderOption{value}
}

// endregion
// region -> service resolver

type serviceResolverOption struct {
	value resolver.ServiceResolver
}

func (o *serviceResolverOption) apply(p *tokenBasedUserDetailsProvider) {
	if o.value != nil {
		p.serviceResolver = o.value
	}
}

func UdpWithServiceResolver(value resolver.ServiceResolver) UserDetailsProviderOption {
	return &serviceResolverOption{value}
}

// endregion
// region -> service resolver

type pathProcessorOption struct {
	value resolver.PathProcessor
}

func (o *pathProcessorOption) apply(p *tokenBasedUserDetailsProvider) {
	if o.value != nil {
		p.pathProcessor = o.value
	}
}

func UdpWithPathProcessor(value resolver.PathProcessor) UserDetailsProviderOption {
	return &pathProcessorOption{value}
}

// endregion
// region -> auth endpoint

type authEndpointOption struct {
	value string
}

func (o *authEndpointOption) apply(p *tokenBasedUserDetailsProvider) {
	if o.value != "" {
		p.endpoint = o.value
	}
}

func UdpWithAuthEndpoint(value string) UserDetailsProviderOption {
	return &authEndpointOption{value}
}

// endregion
// region -> method

type methodOption struct {
	value string
}

func (o *methodOption) apply(p *tokenBasedUserDetailsProvider) {
	if o.value != "" {
		p.method = o.value
	}
}

func UdpWithMethod(value string) UserDetailsProviderOption {
	return &methodOption{value}
}

// endregion
// region -> response parser

type responseParserOption struct {
	value ResponseParser
}

func (o *responseParserOption) apply(p *tokenBasedUserDetailsProvider) {
	if o.value != nil {
		p.responseParser = o.value
	}
}

func UdpWithResponseParser(value ResponseParser) UserDetailsProviderOption {
	return &responseParserOption{value}
}

// endregion

// endregion

// region - provider

type tokenBasedUserDetailsProvider struct {
	authProvider    AuthProvider
	serviceResolver resolver.ServiceResolver
	pathProcessor   resolver.PathProcessor
	endpoint        string
	method          string
	logger          logging.Logger
	responseParser  ResponseParser
}

func NewTokenBasedUserDetailsProvider(options ...UserDetailsProviderOption) UserDetailsProvider {
	udp := &tokenBasedUserDetailsProvider{
		logger: logging.GetLogger("rest-user-details-provider"),
		method: "GET",
	}
	for _, option := range options {
		if option != nil {
			option.apply(udp)
		}
	}
	return udp
}

func (p *tokenBasedUserDetailsProvider) Get(token string) (UserDetails, error) {

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

func (p *tokenBasedUserDetailsProvider) exchange(endpoint, header string) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest(p.method, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("could not generate delegated authorization request: %s", err)
	}
	req.Header.Set(constants.HdrAuthorization, header)
	return client.Do(req)
}
func (p *tokenBasedUserDetailsProvider) processAuthResponse(res *http.Response) (UserDetails, error) {
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

// endregion
