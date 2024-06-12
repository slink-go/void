package rate

import (
	"github.com/slink-go/api-gateway/middleware/constants"
	"regexp"
	"strings"
	"time"
)

// region - options

type CustomLimitOption interface {
	applyCustom(*customRateLimit)
}

// region -> pattern

func WithCustomPattern(value string) CustomLimitOption {
	return &customOptionPattern{
		value: value,
	}
}

type customOptionPattern struct {
	value string
}

func (c *customOptionPattern) applyCustom(r *customRateLimit) {
	r.pattern = c.value
	pattern := strings.ReplaceAll(c.value, "*", ".*")
	r.re = regexp.MustCompile(pattern)
}

// endregion
// region -> period

func WithCustomPeriod(value time.Duration) CustomLimitOption {
	return &customOptionPeriod{
		value: value,
	}
}

type customOptionPeriod struct {
	value time.Duration
}

func (c *customOptionPeriod) applyCustom(r *customRateLimit) {
	r.period = c.value
}

// endregion
//region -> limit

func WithCustomLimit(value int64) CustomLimitOption {
	return &customOptionLimit{
		value: value,
	}
}

type customOptionLimit struct {
	value int64
}

func (c *customOptionLimit) applyCustom(r *customRateLimit) {
	r.limit = c.value
}

// endregion

// endregion
// region - custom limit

func NewCustomRateLimit() *customRateLimit {
	return &customRateLimit{
		pattern: "*",
		re:      regexp.MustCompile(".*"),
		period:  time.Second,
		limit:   constants.MaxInt64,
	}
}

type customRateLimit struct {
	pattern string
	re      *regexp.Regexp
	period  time.Duration
	limit   int64
}

func (c *customRateLimit) GetLimit() int64 {
	return c.limit
}
func (c *customRateLimit) GetPeriod() time.Duration {
	return c.period
}
func (c *customRateLimit) GetPattern() string {
	return c.pattern
}

func (c *customRateLimit) apply(lm *limiterImpl) {
	lm.custom = append(lm.custom, *c)
	//TODO: sort by pattern (?)
}

// endregion
