package security

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
}

// endregion
// region - UserDetails

type UserDetails map[string]string

// endregion
// region - UserDetails

type UserDetailsProvider interface {
	Get(token string) (UserDetails, error)
}

// endregion
