package cache

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCacheSetAndGet(t *testing.T) {
	cache := NewCache[int, string](time.Minute)

	cache.Set(1, "test1")
	value, found := cache.Get(1)
	assert.True(t, found, "Expected to find key 1")
	assert.Equal(t, "test1", value, "Expected value to be 'test1'")

	_, found = cache.Get(2)
	assert.False(t, found, "Expected not to find key 2")
}

func TestCacheCleanup(t *testing.T) {
	ttl := 100 * time.Millisecond
	cache := NewCache[int, string](ttl)

	cache.Set(1, "value1")

	val, ok := cache.Get(1)
	assert.True(t, ok)
	assert.Equal(t, "value1", val)

	//wait for cleanup
	time.Sleep(200 * time.Millisecond)

	//check if value is cleaned up
	val, ok = cache.Get(1)
	assert.False(t, ok)
	assert.Equal(t, "", val)
}

func TestCacheDelete(t *testing.T) {
	cache := NewCache[int, string](time.Minute)

	cache.Set(1, "test1")
	cache.Delete(1)
	_, found := cache.Get(1)
	assert.False(t, found, "Expected not to find key 1 after deletion")
}

func TestCacheClear(t *testing.T) {
	cache := NewCache[int, string](time.Minute)

	cache.Set(1, "test1")
	cache.Set(2, "test2")
	cache.Clear()

	_, found := cache.Get(1)
	assert.False(t, found, "Expected not to find key 1 after clear")
	_, found = cache.Get(2)
	assert.False(t, found, "Expected not to find key 2 after clear")
}

func TestCacheTTLExpiration(t *testing.T) {
	cache := NewCache[int, string](50 * time.Millisecond)

	cache.Set(1, "test1")
	time.Sleep(100 * time.Millisecond)

	_, found := cache.Get(1)
	assert.False(t, found, "Expected not to find key 1 after TTL expiration")
}

func TestCacheStopCleanup(t *testing.T) {
	cache := NewCache[int, string](50 * time.Millisecond)

	cache.Set(1, "test1")
	cache.StopCleanup()
	time.Sleep(100 * time.Millisecond)

	// Since cleanup has been stopped, key 1 should still be present
	value, found := cache.Get(1)
	assert.True(t, found, "Expected to find key 1 after stopping cleanup")
	assert.Equal(t, "test1", value, "Expected value to be 'test1'")
}

func TestCacheConcurrentAccess(t *testing.T) {
	cache := NewCache[int, string](time.Minute)
	wg := sync.WaitGroup{}
	const numGoroutines = 100
	const numItems = 1000

	// Concurrently set values
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			for j := 0; j < numItems; j++ {
				cache.Set(goroutineID*numItems+j, "value")
			}
			wg.Done()
		}(i)
	}

	wg.Wait()

	// Concurrently get values
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			for j := 0; j < numItems; j++ {
				_, found := cache.Get(goroutineID*numItems + j)
				assert.True(t, found, "Expected to find key")
			}
			wg.Done()
		}(i)
	}

	wg.Wait()
}

func TestCacheSetAndGetAsync(t *testing.T) {
	cache := NewCache[int, string](time.Minute)
	var wg sync.WaitGroup

	setDone := make(chan struct{})

	wg.Add(1)
	go func() {
		defer wg.Done()
		cache.Set(1, "test1")
		close(setDone)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-setDone
		value, found := cache.Get(1)
		assert.True(t, found, "Expected to find key 1")
		assert.Equal(t, "test1", value, "Expected value to be 'test1'")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-setDone
		_, found := cache.Get(2)
		assert.False(t, found, "Expected not to find key 2")
	}()

	wg.Wait()
}

func TestCacheDeleteAsync(t *testing.T) {
	cache := NewCache[int, string](time.Minute)
	var wg sync.WaitGroup

	setDone := make(chan struct{})
	deleteDone := make(chan struct{})

	wg.Add(1)
	go func() {
		defer wg.Done()
		cache.Set(1, "test1")
		close(setDone)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-setDone
		cache.Delete(1)
		close(deleteDone)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-deleteDone
		_, found := cache.Get(1)
		assert.False(t, found, "Expected not to find key 1 after deletion")
	}()

	wg.Wait()
}

func TestCacheClearAsync(t *testing.T) {
	cache := NewCache[int, string](time.Minute)
	var wg sync.WaitGroup

	setDone := make(chan struct{})
	clearDone := make(chan struct{})

	wg.Add(1)
	go func() {
		defer wg.Done()
		cache.Set(1, "test1")
		cache.Set(2, "test2")
		close(setDone)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-setDone
		cache.Clear()
		close(clearDone)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-clearDone
		_, found := cache.Get(1)
		assert.False(t, found, "Expected not to find key 1 after clear")
		_, found = cache.Get(2)
		assert.False(t, found, "Expected not to find key 2 after clear")
	}()

	wg.Wait()
}

func TestCacheTTLExpirationAsync(t *testing.T) {
	cache := NewCache[int, string](50 * time.Millisecond)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		cache.Set(1, "test1")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(100 * time.Millisecond)
		_, found := cache.Get(1)
		assert.False(t, found, "Expected not to find key 1 after TTL expiration")
	}()

	wg.Wait()
}

func TestCacheStopCleanupAsync(t *testing.T) {
	cache := NewCache[int, string](50 * time.Millisecond)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		cache.Set(1, "test1")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		cache.StopCleanup()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(100 * time.Millisecond)
		value, found := cache.Get(1)
		assert.True(t, found, "Expected to find key 1 after stopping cleanup")
		assert.Equal(t, "test1", value, "Expected value to be 'test1'")
	}()

	wg.Wait()
}

func TestCacheConcurrentAccessAsync(t *testing.T) {
	cache := NewCache[int, string](time.Minute)
	wg := sync.WaitGroup{}
	const numGoroutines = 100
	const numItems = 1000

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < numItems; j++ {
				cache.Set(goroutineID*numItems+j, "value")
			}
		}(i)
	}

	wg.Wait()

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < numItems; j++ {
				_, found := cache.Get(goroutineID*numItems + j)
				assert.True(t, found, "Expected to find key")
			}
		}(i)
	}

	wg.Wait()
}

func TestCacheConcurrentSetAndDeleteAsync(t *testing.T) {

	c := NewCache[int, string](time.Minute)

	const numGoroutines = 100
	const key = 1
	const value = "value"

	var wg sync.WaitGroup

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			c.Set(key, value)
			wg.Done()
		}()
	}

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			c.Delete(key)
			wg.Done()
		}()
	}

	wg.Wait()

	_, found := c.Get(key)
	assert.False(t, found, "Expected not to find key")
	if found {
		t.Errorf("Expected not to find key, but found")
	}
}
