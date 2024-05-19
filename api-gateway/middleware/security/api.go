package security

// region - Auth

type Type int

const (
	TypeNone Type = iota
	TypeBasic
	TypeBearer
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

type UserDetails struct {
	UserId   string
	UserName string
	UserRole string
}

// endregion
// region - UserDetails

type UserDetailsProvider interface {
	Get(auth Auth) (*UserDetails, error)
}

// endregion
