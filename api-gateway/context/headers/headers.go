package headers

import "strings"

const (
	CtxAuthToken = "Ctx-Auth-Token"
	CtxUserId    = "Ctx-User-Id"
	CtxUserName  = "Ctx-User-Name"
	CtxUserRole  = "Ctx-User-Role"
	CtxLocale    = "Ctx-Locale"
)

func Get(values []string) string {
	if values == nil || len(values) == 0 {
		return ""
	}
	return values[0]
}
func GetAll(values []string, separator string) string {
	if values == nil || len(values) == 0 {
		return ""
	}
	return strings.Join(values, separator)
}
