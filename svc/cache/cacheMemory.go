package cache

import (
	"context"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
)

type (
	Memory struct {
		Cleanup    int
		Expiration int

		cache *cache.Cache

		logger        *logrus.Entry
		context       context.Context
		contextCancel context.CancelFunc

		mu            sync.RWMutex
		isInitialized bool
	}

	MemoryConfigInput struct {
		Expiration int
		Cleanup    int
		Logger     *logrus.Entry
	}
)

func NewCacheMemory(cacheConfig *MemoryConfigInput) ICacheCache {
	c := new(Memory)
	expiration := cacheConfig.Expiration

	if cacheConfig.Expiration == 0 {
		expiration = -1
	}

	c.cache = cache.New(
		time.Duration(expiration)*time.Second,
		time.Duration(cacheConfig.Cleanup)*time.Second,
	)

	c.Expiration = expiration
	c.Cleanup = cacheConfig.Cleanup

	c.logger = cacheConfig.Logger

	return c
}

func (c *Memory) Start() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.logger.
		Infof("starting...")

	listenCtx, listenCancelCtx := context.WithCancel(context.Background())
	c.context = listenCtx
	c.contextCancel = listenCancelCtx

	c.isInitialized = true

	c.logger.
		Infof("starting... done!")
}

func (c *Memory) Flush() {
	c.mu.RLock()
	defer c.mu.RUnlock()

	c.logger.
		Debugf("flushing...")

	if c.cache != nil && c.isInitialized {
		c.cache.Flush()
	}

	c.logger.
		Debugf("flushing... done!")
}

func (c *Memory) Shutdown() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.logger.
		Infof("stopping...")

	if c.cache != nil && c.isInitialized {
		c.contextCancel()
		c.isInitialized = false

		c.cache.Flush()
		c.cache = nil
	}

	c.logger.
		Infof("stopping... done!")
}

func (c *Memory) Get(key string) (*Item, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.cache != nil && c.isInitialized {
		if item, ok := c.cache.Get(key); ok {
			return item.(*Item), ok
		}
	}

	return nil, false
}

func (c *Memory) Replace(key string, value *Item, d time.Duration) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if d >= 0 && c.cache != nil && c.isInitialized {
		if d == 0 {
			d = -1
		}

		return c.cache.Replace(key, value, d)
	}

	return nil
}

func (c *Memory) ReplaceIfExists(key string, value *Item, d time.Duration) error {
	if _, ok := c.Get(key); ok {
		return c.Replace(key, value, d)
	}

	return c.Add(key, value, d)
}

func (c *Memory) Delete(key string) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.cache != nil && c.isInitialized {
		c.cache.Delete(key)
	}
}

func (c *Memory) Add(key string, value *Item, d time.Duration) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if d >= 0 && c.cache != nil && c.isInitialized {
		if d == 0 {
			d = -1
		}

		return c.cache.Add(key, value, d)
	}

	return nil
}

func (c *Memory) IsEnabled() bool {
	if c.cache != nil && c.isInitialized {
		return c.Expiration > -1
	}

	return false
}

func (c *Memory) IsDisabled() bool {
	return !c.IsEnabled()
}
