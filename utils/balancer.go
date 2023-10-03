package util

import (
	"math/rand"
	"sync"
	"time"

	"github.com/coocood/freecache"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type sameNodeBalancer struct {
	targets []*middleware.ProxyTarget
	mutex   sync.Mutex
	random  *rand.Rand
	store   freecache.Cache
	TTL     int
}

func NewSameNodeBalancer(targets []*middleware.ProxyTarget, ttl int) middleware.ProxyBalancer {
	b := sameNodeBalancer{}
	b.targets = targets
	b.random = rand.New(rand.NewSource(int64(time.Now().Nanosecond())))
	b.store = *freecache.NewCache(1024 * 1024 * 1)
	b.TTL = ttl
	return &b
}

func (b *sameNodeBalancer) AddTarget(target *middleware.ProxyTarget) bool {
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

func (b *sameNodeBalancer) RemoveTarget(name string) bool {
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

func (b *sameNodeBalancer) Next(c echo.Context) *middleware.ProxyTarget {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	if len(b.targets) == 0 {
		return nil
	} else if len(b.targets) == 1 {
		return b.targets[0]
	}

	var i int
	ipKey := []byte(c.RealIP())

	got, err := b.store.Get(ipKey)
	if err != nil {
		i = b.random.Intn(len(b.targets))
		b.store.Set(ipKey, []byte{byte(i)}, 10)
	} else {
		i = int(got[0])
	}

	return b.targets[i]
}
