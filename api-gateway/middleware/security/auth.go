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
// region - Provider

func NewHttpHeaderAuthProvider() AuthProvider {
	return &httpHeaderAuthProvider{
		logger: logging.GetLogger("header-auth-provider"),
	}
}

type httpHeaderAuthProvider struct {
	logger logging.Logger
}

func (ap *httpHeaderAuthProvider) Get(args ...string) (Auth, error) {
	if args == nil || len(args) == 0 || args[0] == "" {
		return NewNoAuth(), nil
	}
	header := args[0]
	if strings.HasPrefix(header, bearerPrefix) {
		token := strings.Replace(header, bearerPrefix, "", -1)
		return NewTokenAuth(strings.TrimSpace(token)), nil
	} else if strings.HasPrefix(header, basicPrefix) {
		login, password, ok := ap.parseBasicAuth(header)
		if !ok {
			return nil, fmt.Errorf("could not parse basic auth header")
		}
		return NewBasicAuth(login, password), nil
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
