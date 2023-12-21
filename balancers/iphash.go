package balancers

import (
	"math/rand"
	"sync"
	"time"

	echocache "github.com/fraidev/go-echo-cache"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type ipHashBalancer struct {
	targets     []*middleware.ProxyTarget
	retryTarget *middleware.ProxyTarget
	mutex       sync.Mutex
	random      *rand.Rand
	store       echocache.Cache
	TTL         int
}

func NewIPHashBalancer(targets []*middleware.ProxyTarget, retryTarget *middleware.ProxyTarget, ttl int, store echocache.Cache) middleware.ProxyBalancer {
	b := ipHashBalancer{}
	b.targets = targets
	b.retryTarget = retryTarget
	b.random = rand.New(rand.NewSource(int64(time.Now().Nanosecond())))
	b.store = store
	b.TTL = ttl
	return &b
}

func (b *ipHashBalancer) AddTarget(target *middleware.ProxyTarget) bool {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	for _, t := range b.targets {
		if t.Name == target.Name {
			return false
		}
	}
	b.targets = append(b.targets, target)
	return true
}

func (b *ipHashBalancer) RemoveTarget(name string) bool {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	for i, t := range b.targets {
		if t.Name == name {
			b.targets = append(b.targets[:i], b.targets[i+1:]...)
			return true
		}
	}
	return false
}

func (b *ipHashBalancer) Next(c echo.Context) *middleware.ProxyTarget {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if len(b.targets) == 0 {
		return nil
	} else if len(b.targets) == 1 {
		return b.targets[0]
	}

	if c.Get("retry") != nil {
		return b.retryTarget
	}

	ctx := c.Request().Context()
	ip := []byte(c.RealIP())
	got, err := b.store.Get(ctx, ip)
	if err != nil {
		i := b.random.Intn(len(b.targets))
		b.store.Set(ctx, ip, []byte{byte(i)}, b.TTL)
		return b.targets[i]
	}

	return b.targets[int(got[0])]
}
