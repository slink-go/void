package context

import (
	"fmt"
)

func (bcp *basicContextProvider) processAuthOption(options []Option) (map[string][]string, error) {
	if bcp.authProvider == nil {
		return nil, nil
	}
	if bcp.userDetailsProvider == nil {
		return nil, nil
	}
	if options == nil || len(options) == 0 {
		return nil, nil
	}
	if option := findAuthOption(options...); option != nil {
		auth, err := bcp.authProvider.Get(option.Value())
		if err != nil {
			return nil, fmt.Errorf("could not get user details from auth provider: %v", err)
		}
		userDetails, err := bcp.userDetailsProvider.Get(auth)
		if err != nil {
			return nil, fmt.Errorf("could not get user details from auth provider: %v", err)
		}
		if userDetails == nil {
			return nil, fmt.Errorf("got nil user details from auth provider: %v", err)
		}
		result := make(map[string][]string)
		if userDetails.UserId != "" {
			result[CtxUserId] = []string{userDetails.UserId}
		}
		if userDetails.UserName != "" {
			result[CtxUserName] = []string{userDetails.UserName}
		}
		if userDetails.UserRole != "" {
			result[CtxUserRole] = []string{userDetails.UserRole}
		}
		return result, nil
	}
	return nil, nil
}

func (bcp *basicContextProvider) processLocaleOption(options []Option) map[string][]string {
	if options == nil || len(options) == 0 {
		return nil
	}
	if option := findLangParamOption(options...); option != nil {
		if option.Value() != "" {
			return headerMap(CtxLocale, option.Value())
		}
	}
	if option := findAcceptHeaderOption(options...); option != nil {
		return headerMap(CtxLocale, option.Value())
	}
	return nil
}

func findAuthOption(options ...Option) Option {
	for _, option := range options {
		if _, ok := option.(*AuthContextOption); ok {
			return option
		}
	}
	return nil
}
func findAcceptHeaderOption(options ...Option) Option {
	for _, option := range options {
		if _, ok := option.(*LocalizationOption); ok {
			return option
		}
	}
	return nil
}
func findLangParamOption(options ...Option) Option {
	for _, option := range options {
		if _, ok := option.(*LangParamOption); ok {
			return option
		}
	}
	return nil
}
func headerMap(key, value string) map[string][]string {
	if value == "" {
		return nil
	}
	result := make(map[string][]string)
	result[key] = []string{value}
	return result
}
