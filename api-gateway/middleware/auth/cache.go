package auth

import (
	"context"
	"github.com/jellydator/ttlcache/v3"
	"github.com/slink-go/api-gateway/middleware/security"
	"github.com/slink-go/logging"
	"math"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Cache interface {
	Get(token string) (security.UserDetails, bool)
	Set(token string, user security.UserDetails)
}

func NewUserDetailsCache(ttl time.Duration) Cache {
	l := logging.GetLogger("user-details-cache")
	cache := ttlcache.New[string, security.UserDetails](
		ttlcache.WithTTL[string, security.UserDetails](ttl),
	)
	cache.OnInsertion(func(ctx context.Context, item *ttlcache.Item[string, security.UserDetails]) {
		l.Trace("insert: %v %v", keyLog(item.Key()), item.ExpiresAt())
	})
	cache.OnEviction(func(ctx context.Context, reason ttlcache.EvictionReason, item *ttlcache.Item[string, security.UserDetails]) {
		if reason == ttlcache.EvictionReasonExpired {
			l.Trace("evict: %v", keyLog(item.Key()))
		}
	})
	udc := userDetailsCache{
		logger: l,
		cache:  cache,
	}
	go udc.run(ttl / 2)
	go udc.handleSignal()
	return &udc
}

type userDetailsCache struct {
	cache   *ttlcache.Cache[string, security.UserDetails]
	stopChn chan struct{}
	logger  logging.Logger
}

func (c *userDetailsCache) Get(token string) (user security.UserDetails, ok bool) {
	item := c.cache.Get(token)
	if item == nil {
		c.logger.Trace("not found: %v", keyLog(token))
		return nil, false
	}
	v := item.Value()
	if v == nil {
		c.logger.Trace("no value for: %v", keyLog(token))
		return nil, false
	}
	c.logger.Trace("found cached value for: %v", keyLog(token))
	return v, true
}
func (c *userDetailsCache) Set(token string, user security.UserDetails) {
	c.logger.Trace("set: %v", keyLog(token))
	c.cache.Set(token, user, ttlcache.DefaultTTL)
}

func (c *userDetailsCache) run(duration time.Duration) {
	v := duration.Milliseconds()
	vv := time.Duration(math.Max(1000, float64(v)/2))
	delay := time.Millisecond * vv
	timer := time.NewTimer(delay)
	for {
		select {
		case <-c.stopChn:
			return
		case <-timer.C:
			c.cache.DeleteExpired()
		}
		timer.Reset(delay)
	}
}

func (c *userDetailsCache) handleSignal() {
	sigChn := make(chan os.Signal)
	signal.Notify(sigChn, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
	for {
		switch <-sigChn {
		case syscall.SIGINT:
			fallthrough
		case syscall.SIGKILL:
			fallthrough
		case syscall.SIGTERM:
			c.logger.Info("shutdown user details cache")
			c.stopChn <- struct{}{}
			return
		}
	}
}

func keyLog(token string) string {
	v := token[:3] + "..."
	return v + token[len(token)-3:]
}
