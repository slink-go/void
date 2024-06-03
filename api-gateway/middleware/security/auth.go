package security

import (
	"encoding/base64"
	"fmt"
	"github.com/slink-go/logging"
	"strings"
)

const bearerPrefix = "Bearer "
const basicPrefix = "Basic "

// region - None

func NewNoAuth() Auth {
	return &noAuth{}
}

type noAuth struct {
}

func (a *noAuth) GetType() Type {
	return TypeNone
}
func (a *noAuth) GetValue() interface{} {
	return nil
}

// endregion
// region - Basic

func NewBasicAuth(login, password string) Auth {
	return &basicAuth{
		login:    login,
		password: password,
	}
}

type basicAuth struct {
	login    string
	password string
}

func (a *basicAuth) GetType() Type {
	return TypeBasic
}
func (a *basicAuth) GetValue() interface{} {
	return []string{a.login, a.password}
}

// endregion
// region - Token

func NewTokenAuth(token string) Auth {
	return &tokenAuth{
		token: token,
	}
}

type tokenAuth struct {
	token string
}

func (a *tokenAuth) GetType() Type {
	return TypeBearer
}
func (a *tokenAuth) GetValue() interface{} {
	return a.token
}

// endregion
// region - Cookie

func NewCookieAuth(token string) Auth {
	return &cookieAuth{
		token: token,
	}
}

type cookieAuth struct {
	token string
}

func (a *cookieAuth) GetType() Type {
	return TypeCookie
}
func (a *cookieAuth) GetValue() interface{} {
	return a.token
}

// endregion

// region - Http Header Auth Provider

func NewHttpHeaderAuthProvider() AuthProvider {
	return &httpHeaderAuthProvider{
		logger: logging.GetLogger("header-auth-provider"),
	}
}
func (ap *httpHeaderAuthProvider) WithProvider(provider AuthProvider) AuthProvider {
	return ap
}

type httpHeaderAuthProvider struct {
	logger logging.Logger
}

func (ap *httpHeaderAuthProvider) Get(args ...string) (Auth, error) {
	if args == nil || len(args) == 0 {
		return NewNoAuth(), nil
	}
	for _, header := range args {
		if strings.HasPrefix(header, bearerPrefix) {
			token := strings.TrimSpace(header[len(bearerPrefix):])
			if isValidToken(token) {
				return NewTokenAuth(strings.TrimSpace(token)), nil
			} else {
				ap.logger.Warning("invalid token: %s", token)
				continue
			}
		} else if strings.HasPrefix(header, basicPrefix) {
			login, password, ok := ap.parseBasicAuth(header)
			if !ok {
				ap.logger.Warning("could not parse basic auth header")
				continue
			}
			return NewBasicAuth(login, password), nil
		}
	}
	return nil, fmt.Errorf("could not parse auth header")
}

func (ap *httpHeaderAuthProvider) parseBasicAuth(auth string) (username, password string, ok bool) {
	c, err := base64.StdEncoding.DecodeString(auth[len(basicPrefix):])
	if err != nil {
		return "", "", false
	}
	cs := string(c)
	username, password, ok = strings.Cut(cs, ":")
	if !ok {
		return "", "", false
	}
	return username, password, true
}

// endregion
// region - Cookie Auth Provider

func NewCookieAuthProvider() AuthProvider {
	return &cookieAuthProvider{
		logger: logging.GetLogger("cookie-auth-provider"),
	}
}
func (ap *cookieAuthProvider) WithProvider(provider AuthProvider) AuthProvider {
	return ap
}

type cookieAuthProvider struct {
	logger logging.Logger
}

func (ap *cookieAuthProvider) Get(args ...string) (Auth, error) {
	if args == nil || len(args) == 0 {
		return NewNoAuth(), nil
	}
	for _, token := range args {
		if isValidToken(token) {
			return NewCookieAuth(token), nil
		}
	}
	return nil, fmt.Errorf("could not parse auth cookie")
}

// endregion
// region - Auth Providers Chain

func NewAuthChain() AuthProvider {
	return &authChain{
		logger: logging.GetLogger("auth-chain"),
	}
}
func (p *authChain) WithProvider(provider AuthProvider) AuthProvider {
	p.aps = append(p.aps, provider)
	return p
}

type authChain struct {
	logger logging.Logger
	aps    []AuthProvider
}

func (ac *authChain) Get(args ...string) (Auth, error) {
	for _, ap := range ac.aps {
		auth, err := ap.Get(args...)
		if err != nil {
			ac.logger.Debug("auth error: %s", err)
		}
		if err == nil && auth != nil && auth.GetType() != TypeNone {
			return auth, nil
		}
	}
	return NewNoAuth(), nil
}

// endregion

// region - common

func isValidToken(input string) bool {
	// TODO: проверить, что токен валидный
	return input != ""
}

//endregion
