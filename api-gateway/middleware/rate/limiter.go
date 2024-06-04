package rate

import (
	"github.com/slink-go/api-gateway/cmd/common/env"
	"github.com/slink-go/api-gateway/middleware"
	"github.com/slink-go/api-gateway/middleware/constants"
	"github.com/slink-go/logging"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"
	"time"
)

// https://adam-p.ca/blog/2022/03/x-forwarded-for/
// https://github.com/realclientip/realclientip-go
// https://github.com/realclientip/realclientip-go/tree/main/_examples
// https://github.com/realclientip/realclientip-go/wiki/Single-IP-Headers

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
		option.applyCustom(lm)
	}
	return lm
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
		option.apply(&lm)
	}
	if lm.store == nil {
		panic("rate-limiter store is not set")
	}
	return &lm
}

type Limiter interface {
	Get(url string) *limiter.Limiter
}

type limiterImpl struct {
	global struct {
		period time.Duration
		limit  int64
	}
	custom []customRateLimit
	store  limiter.Store
	logger logging.Logger
}

func (l *limiterImpl) Get(url string) *limiter.Limiter {
	if l.store == nil {
		l.logger.Error("rate limit store not set")
		return nil
	}
	rate := l.getRate(url)
	lm := limiter.New(l.store, rate)
	return lm
}

func (l *limiterImpl) getRate(url string) limiter.Rate {
	if rate, ok := l.getCustomRate(url); ok {
		return rate
	} else {
		return limiter.Rate{
			Limit:  l.global.limit,
			Period: l.global.period,
		}
	}
}
func (l *limiterImpl) getCustomRate(url string) (limiter.Rate, bool) {
	for _, crl := range l.custom {
		if middleware.Match(url, crl.pattern) {
			return limiter.Rate{
				Period: crl.period,
				Limit:  crl.limit,
			}, true
		}
	}
	return limiter.Rate{}, false
}

// endregion
