package rate

import "math"

type Limiter interface {
	GetLimit() int
}

type rpsLimiter struct {
	limit int
}

func NewLimiter(limit int) Limiter {
	l := limit
	if l <= 0 {
		l = math.MaxInt64
	}
	return &rpsLimiter{
		limit: l,
	}
}

func (l *rpsLimiter) GetLimit() int {
	return l.limit
}
