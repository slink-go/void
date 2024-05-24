package rate

const MaxUint = ^uint(0)
const MinUint = 0
const MaxInt = int(MaxUint >> 1)
const MinInt = -MaxInt - 1

type Limiter interface {
	GetLimit() int
}

type rpsLimiter struct {
	limit int
}

func NewLimiter(limit int) Limiter {
	l := limit
	if l <= 0 {
		l = MaxInt
	}
	return &rpsLimiter{
		limit: l,
	}
}

func (l *rpsLimiter) GetLimit() int {
	return l.limit
}
