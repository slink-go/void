package rate

type Limiter interface {
	GetLimit() int
}

type rpsLimiter struct {
	limit int
}

func NewLimiter(limit int) Limiter {
	return &rpsLimiter{
		limit: limit,
	}
}

func (l *rpsLimiter) GetLimit() int {
	return l.limit
}
