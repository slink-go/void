package context

import (
	"github.com/slink-go/api-gateway/middleware/security"
)

type Provider interface {
	WithAuthProvider(ap security.AuthProvider) Provider
	WithUserDetailsProvider(udp security.UserDetailsProvider) Provider
	GetContext(options ...Option) map[string][]string
}

type Option interface {
	Value() string
}

func NewAuthContextOption(value string) Option {
	return &AuthContextOption{
		authHeaderValue: value,
	}
}

type AuthContextOption struct {
	authHeaderValue string
}

func (o *AuthContextOption) Value() string {
	return o.authHeaderValue
}

func NewLocalizationOption(value string) Option {
	return &LocalizationOption{
		acceptLangHeaderValue: value,
	}
}

type LocalizationOption struct {
	acceptLangHeaderValue string
}

func (o *LocalizationOption) Value() string {
	return o.acceptLangHeaderValue
}

func NewLangParamOption(value string) Option {
	return &LangParamOption{
		langQueryParamValue: value,
	}
}

type LangParamOption struct {
	langQueryParamValue string
}

func (o *LangParamOption) Value() string {
	return o.langQueryParamValue
}
