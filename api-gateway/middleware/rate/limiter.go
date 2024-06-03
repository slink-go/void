package rate

import "github.com/slink-go/api-gateway/middleware/constants"

type Limiter interface {
	GetLimit() int
}

type rpsLimiter struct {
	limit int
}

func NewLimiter(limit int) Limiter {
	l := limit
	if l <= 0 {
		l = constants.MaxInt
	}
	return &rpsLimiter{
		limit: l,
	}
}

func (l *rpsLimiter) GetLimit() int {
	return l.limit
}
