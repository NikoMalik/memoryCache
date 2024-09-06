package cache

import (
	"time"
	"unsafe"

	"github.com/alphadose/haxmap"
)

const (
	iter0       = 1 << 3
	elementNum0 = 1 << 10
)

type Signed interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

type Unsigned interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

type Integer interface {
	Signed | Unsigned
}

type Float interface {
	~float32 | ~float64
}

type Complex interface {
	~complex64 | ~complex128
}

type Ordered interface {
	Integer | Float | ~string
}

type hashable interface {
	Integer | Float | Complex | ~string | uintptr | ~unsafe.Pointer
}

type CachedItem[V any] struct {
	Value       V
	CreatedTime time.Time
}

type Cache[T hashable, V any] struct {
	cache       *haxmap.Map[T, *CachedItem[V]]
	ttl         time.Duration
	stopCleanup chan struct{}
}

func NewCache[T hashable, V any](ttl time.Duration) *Cache[T, V] {
	c := &Cache[T, V]{
		cache:       haxmap.New[T, *CachedItem[V]](iter0 * elementNum0),
		ttl:         ttl,
		stopCleanup: make(chan struct{}),
	}
	go c.startCleanupRoutine()
	return c
}

func (c *Cache[T, V]) Set(key T, value V) {

	c.cache.Set(key, &CachedItem[V]{
		Value:       value,
		CreatedTime: time.Now(),
	})
}

func (c *Cache[T, V]) Get(key T) (V, bool) {
	val, ok := c.cache.Get(key)
	if !ok {

		var zero V
		return zero, false
	}
	item := val

	return item.Value, true
}

func (c *Cache[T, V]) Delete(key T) {
	c.cache.Del(key)
}

func (c *Cache[T, V]) Clear() {
	c.cache.ForEach(func(key T, value *CachedItem[V]) bool {
		c.cache.Del(key)
		return true
	})
}

func (c *Cache[T, V]) startCleanupRoutine() {
	ticker := time.NewTicker(c.ttl)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			c.cleanup()
		case <-c.stopCleanup:
			return
		}
	}
}

func (c *Cache[T, V]) cleanup() {
	now := time.Now()
	c.cache.ForEach(func(key T, value *CachedItem[V]) bool {
		if now.Sub(value.CreatedTime) > c.ttl {
			c.cache.Del(key)
		}
		return true
	})
}

func (c *Cache[T, V]) StopCleanup() {
	close(c.stopCleanup)
}
