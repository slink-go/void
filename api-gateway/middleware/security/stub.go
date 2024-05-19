package security

import "fmt"

func NewStubUserDetailsProvider() UserDetailsProvider {
	return &stubUserDetailsProvider{}
}

type stubUserDetailsProvider struct {
}

func (p *stubUserDetailsProvider) Get(auth Auth) (*UserDetails, error) {
	if auth == nil {
		return nil, fmt.Errorf("nil auth received")
	}
	switch auth.GetType() {
	case TypeNone:
		return &UserDetails{}, nil
	case TypeBasic:
		v := auth.GetValue().([]string)
		return &UserDetails{"stub-user-id", v[0], "USER"}, nil
	case TypeBearer:
		return &UserDetails{"stub-user-id", "stub-user-name", "USER"}, nil
	}
	return nil, fmt.Errorf("unsupported auth type: %v", auth.GetType())
}
