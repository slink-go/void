package context

import (
	"github.com/slink-go/api-gateway/middleware/rate"
	"github.com/slink-go/api-gateway/middleware/security"
	"github.com/slink-go/logging"
)

// region - basicContextProvider

type basicContextProvider struct {
	authProvider        security.AuthProvider
	userDetailsProvider security.UserDetailsProvider
	rateLimiter         rate.Limiter
	localeProvider      func(args ...interface{}) string
	logger              logging.Logger
}

func CreateContextProvider() Provider {
	return &basicContextProvider{
		logger: logging.GetLogger("basic-context-provider"),
	}
}

func (bcp *basicContextProvider) WithAuthProvider(ap security.AuthProvider) Provider {
	bcp.authProvider = ap
	return bcp
}
func (bcp *basicContextProvider) WithUserDetailsProvider(udp security.UserDetailsProvider) Provider {
	bcp.userDetailsProvider = udp
	return bcp
}

func (bcp *basicContextProvider) GetContext(options ...Option) map[string][]string {

	result := make(map[string][]string)
	auth, err := bcp.processAuthOption(options)
	if err != nil {
		bcp.logger.Warning("auth processing error: %s", err)
	}
	bcp.mapAppend(result, auth)
	bcp.mapAppend(result, bcp.processLocaleOption(options))

	return result
}

func (bcp *basicContextProvider) mapAppend(mapping map[string][]string, other map[string][]string) {
	if other == nil {
		return
	}
	for k, list := range other {
		if _, ok := mapping[k]; !ok {
			mapping[k] = make([]string, 0)
		}
		for _, value := range list {
			mapping[k] = append(mapping[k], value)
		}
	}
}

// endregion
