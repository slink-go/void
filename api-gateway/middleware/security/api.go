package security

import "github.com/slink-go/api-gateway/resolver"

// region - Auth

type Type int

const (
	TypeNone Type = iota
	TypeBasic
	TypeBearer
	TypeCookie
)

type Auth interface {
	GetType() Type
	GetValue() interface{}
}

// endregion
// region - Provider

type AuthProvider interface {
	Get(args ...string) (Auth, error)
	WithProvider(provider AuthProvider) AuthProvider
}

// endregion
// region - UserDetails

type UserDetails map[string]string

// endregion
// region - UserDetails

type UserDetailsProvider interface {
	Get(token string) (UserDetails, error)
	WithAuthProvider(provider AuthProvider) UserDetailsProvider
	WithServiceResolver(resolver resolver.ServiceResolver) UserDetailsProvider
	WithPathProcessor(processor resolver.PathProcessor) UserDetailsProvider
	WithAuthEndpoint(endpoint string) UserDetailsProvider
	WithMethod(method string) UserDetailsProvider
	WithResponseParser(parser ResponseParser) UserDetailsProvider
}

// endregion
