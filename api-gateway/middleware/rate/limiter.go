package rate

import (
	"github.com/slink-go/api-gateway/cmd/common/env"
	"github.com/slink-go/api-gateway/middleware"
	"github.com/slink-go/api-gateway/middleware/constants"
	"github.com/slink-go/logging"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"
	"strings"
	"time"
)

// https://adam-p.ca/blog/2022/03/x-forwarded-for/
// https://github.com/realclientip/realclientip-go
// https://github.com/realclientip/realclientip-go/tree/main/_examples
// https://github.com/realclientip/realclientip-go/wiki/Single-IP-Headers

// region - mode

type LimiterMode int

const (
	LimiterModeUnknown LimiterMode = iota
	LimiterModeOff
	LimiterModeDeny
	LimiterModeDelay
)

var (
	modeTypeNames = map[LimiterMode]string{
		LimiterModeUnknown: "",
		LimiterModeOff:     "OFF",
		LimiterModeDeny:    "DENY",
		LimiterModeDelay:   "DELAY",
	}
	modeTypeValues = map[string]LimiterMode{
		"":      LimiterModeUnknown,
		"OFF":   LimiterModeOff,
		"DENY":  LimiterModeDeny,
		"DELAY": LimiterModeDelay,
	}
)

func (m LimiterMode) String() string {
	return modeTypeNames[m]
}
func parseLimiterMode(s string) LimiterMode {
	s = strings.TrimSpace(strings.ToUpper(s))
	value, ok := modeTypeValues[s]
	if !ok {
		return LimiterModeUnknown
	}
	return value
}

// endregion
// region - options

type Option interface {
	apply(*limiterImpl)
}

// region -> limit

func WithLimit(value int64) Option {
	return &limitOption{
		value: value,
	}
}

type limitOption struct {
	value int64
}

func (o *limitOption) apply(lm *limiterImpl) {
	lm.global.limit = o.value
}

// endregion
// region -> period

func WithPeriod(value time.Duration) Option {
	return &periodOption{
		value: value,
	}
}

type periodOption struct {
	value time.Duration
}

func (o *periodOption) apply(lm *limiterImpl) {
	lm.global.period = o.value
}

// endregion
//region -> store

func WithStore(value limiter.Store) Option {
	return &storeOption{
		value: value,
	}
}
func WithInMemStore() Option {
	store := memory.NewStoreWithOptions(
		limiter.StoreOptions{
			Prefix:          "default:",
			CleanUpInterval: env.DurationOrDefault(env.RateLimitCacheCleanupInterval, time.Second*30),
		},
	)
	return &storeOption{
		value: store,
	}
}

type storeOption struct {
	value limiter.Store
}

func (c *storeOption) apply(r *limiterImpl) {
	r.store = c.value
}

// endregion
// region -> custom

func WithCustom(options ...CustomLimitOption) Option {
	lm := NewCustomRateLimit()
	for _, option := range options {
		if option != nil {
			option.applyCustom(lm)
		}
	}
	return lm
}

// endregion
// region -> mode

func WithMode(value string) Option {
	m := parseLimiterMode(value)
	if m == LimiterModeUnknown {
		m = LimiterModeOff
	}
	return &modeOption{
		value: m,
	}
}

type modeOption struct {
	value LimiterMode
}

func (c *modeOption) apply(r *limiterImpl) {
	r.mode = c.value
}

// endregion

// endregion
// region - limiter

func NewLimiter(options ...Option) Limiter {
	lm := limiterImpl{
		global: struct {
			period time.Duration
			limit  int64
		}{
			period: time.Second,
			limit:  constants.MaxInt64,
		},
		logger: logging.GetLogger("rate-limiter"),
	}
	for _, option := range options {
		if option != nil {
			option.apply(&lm)
		}
	}
	if lm.store == nil {
		panic("rate-limiter store is not set")
	}
	return &lm
}

type Limiter interface {
	Get(url string) *limiter.Limiter
	Mode() LimiterMode
	KeyForPath(path string) string
}

type limiterImpl struct {
	global struct {
		period time.Duration
		limit  int64
	}
	custom []customRateLimit
	store  limiter.Store
	mode   LimiterMode
	logger logging.Logger
}

func (l *limiterImpl) Get(path string) *limiter.Limiter {
	if l.store == nil {
		l.logger.Error("rate limit store not set")
		return nil
	}
	rate := l.getRate(path)
	lm := limiter.New(l.store, rate)
	return lm
}
func (l *limiterImpl) Mode() LimiterMode {
	return l.mode
}
func (l *limiterImpl) KeyForPath(path string) string {
	for _, custom := range l.custom {
		if middleware.Match(path, custom.pattern, custom.re) {
			return custom.pattern
		}
	}
	return "default"
}

func (l *limiterImpl) getRate(path string) limiter.Rate {
	if rate, ok := l.getCustomRate(path); ok {
		return rate
	} else {
		return limiter.Rate{
			Limit:  l.global.limit,
			Period: l.global.period,
		}
	}
}
func (l *limiterImpl) getCustomRate(path string) (limiter.Rate, bool) {
	for _, crl := range l.custom {
		if middleware.Match(path, crl.pattern, crl.re) {
			return limiter.Rate{
				Period: crl.period,
				Limit:  crl.limit,
			}, true
		}
	}
	return limiter.Rate{}, false
}

// endregion
