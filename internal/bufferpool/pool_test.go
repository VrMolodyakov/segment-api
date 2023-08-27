package bufferpool

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuffer(t *testing.T) {
	pool := New()

	buffer := pool.Get()
	for i := 0; i < 5; i++ {
		if !assert.Equal(t, &bytes.Buffer{}, buffer, `should be empty buffer`) {
			return
		}
		pool.Release(buffer)
	}
}
