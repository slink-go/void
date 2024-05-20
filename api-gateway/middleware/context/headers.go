package context

import "strings"

const (
	CtxAuthToken   = "Ctx-Auth-Token"
	CtxUserId      = "Ctx-User-Id"
	CtxUserName    = "Ctx-User-Name"
	CtxUserRole    = "Ctx-User-Role"
	CtxLocale      = "Ctx-Locale"
	CtxProxyTarget = "Ctx-Proxy-Target"
)

func GetHeader(values []string) string {
	if values == nil || len(values) == 0 {
		return ""
	}
	return values[0]
}
func GetHeaders(values []string, separator string) string {
	if values == nil || len(values) == 0 {
		return ""
	}
	return strings.Join(values, separator)
}
