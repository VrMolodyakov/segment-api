package cache

import (
	"sync"
	"time"
)

type Cache[K comparable, V any] struct {
	values  sync.Map
	cleaner cleaner
}

func New[K comparable, V any](cleanUpInterval time.Duration) *Cache[K, V] {
	cleaner := cleaner{
		interval: cleanUpInterval,
		stop:     make(chan struct{}),
	}

	cache := &Cache[K, V]{
		cleaner: cleaner,
	}

	return cache
}

func (ch *Cache[K, V]) Set(key K, value V, expireAt time.Duration) V {
	var expire int64
	if expireAt > 0 {
		expire = time.Now().Add(expireAt).UnixNano()
	}
	item := NewItem[V](value, expire)
	ch.values.Store(key, item)
	return value
}

func (ch *Cache[K, V]) Get(key K) (V, bool) {
	data, found := ch.values.Load(key)
	if !found {
		var emptyValue V
		return emptyValue, false
	}
	item := data.(*Item[V])
	if item.ExpireAt < time.Now().UnixNano() {
		var emptyValue V
		return emptyValue, false
	}

	return item.Value, true
}

func (ch *Cache[K, V]) Delete(key K) {
	ch.values.Delete(key)
}

func (ch *Cache[K, V]) Clean() {
	if ch.cleaner.interval > 0 {
		go ch.clean()
	}
}

func (ch *Cache[K, V]) clean() {
	ticker := time.NewTicker(ch.cleaner.interval)
	for {
		select {
		case <-ticker.C:
			ch.purge()
		case <-ch.cleaner.stop:
			ticker.Stop()
			return
		}
	}
}

func (ch *Cache[K, V]) purge() {
	now := time.Now().UnixNano()
	ch.values.Range(func(key, value interface{}) bool {
		item := value.(*Item[V])
		if item.ExpireAt < now {
			ch.values.Delete(key)
		}
		return true
	})
}

func (ch *Cache[K, V]) Close() {
	ch.cleaner.stopCleaner()
}
