package random

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShiftGeenrator(t *testing.T) {
	gen := NewRandomGenerator(100)
	expected := make([]int, 100)
	for i := range expected {
		expected[i] = 2
	}
	got := make([]int, 100)
	for i := 0; i < 200; i++ {
		got[gen.Next()-1]++
	}

	assert.Equal(t, expected, got)
}

func TestGorutinesGetRandomNumberAndNo2EqualNumbers(t *testing.T) {
	cgen := NewRandomGenerator(1000)
	ctx, cancel := context.WithCancel(context.Background())
	stream := RangeRandomStream(ctx, cgen)
	emitter := NewEmitter(stream)
	got := make([]int, 1000)
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			val := emitter.Next()
			got[val-1]++
			wg.Done()
		}()
	}

	expected := make([]int, 1000)
	for i := range expected {
		expected[i] = 1
	}

	wg.Wait()
	cancel()

	assert.Equal(t, emitter.Next(), 0)
	assert.Equal(t, expected, got)
}
