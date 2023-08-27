package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSet(t *testing.T) {
	cache := New[string, string](30 * time.Second)
	item := cache.Set("abc", "d", 10*time.Second)
	expected := &Item[string]{Value: "d"}
	assert.Equal(t, expected.Value, item)
}

func TestGet(t *testing.T) {
	cache := New[string, string](30 * time.Second)
	_ = cache.Set("abc", "d", 10*time.Second)
	actual, inCache := cache.Get("abc")
	expected := "d"
	assert.Equal(t, true, inCache)
	assert.Equal(t, expected, actual)
}

func TestExpiration(t *testing.T) {
	cache := New[string, string](1 * time.Millisecond)
	_ = cache.Set("abc", "d", 10*time.Millisecond)
	time.Sleep(15 * time.Millisecond)
	actual, inCache := cache.Get("abc")
	assert.Equal(t, false, inCache)
	assert.Equal(t, "", actual)

}

func TestDelete(t *testing.T) {
	cache := New[string, string](1 * time.Second)
	_ = cache.Set("abc", "d", 10*time.Second)
	cache.Delete("abc")
	_, inCache := cache.Get("abc")
	assert.False(t, inCache)
}
