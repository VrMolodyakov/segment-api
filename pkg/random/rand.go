package random

import (
	"context"
	"math/rand"
	"time"
)

type shiftGenerator struct {
	seq    []int
	length int
}

func NewRandomGenerator(length int) *shiftGenerator {
	rand.Seed(time.Now().UnixNano())
	outputSequence := make([]int, length)
	for i := range outputSequence {
		outputSequence[i] = i + 1
	}
	return &shiftGenerator{
		seq:    outputSequence,
		length: length,
	}
}

func (c *shiftGenerator) Next() int {
	i := rand.Intn(c.length)
	val := c.seq[i]
	c.seq[i], c.seq[c.length-1] = c.seq[c.length-1], c.seq[i]
	c.length--
	if c.length == 0 {
		c.length = len(c.seq)
	}
	return val
}

type cycledRandomRange struct {
	stream <-chan int
}

func NewEmitter(stream <-chan int) *cycledRandomRange {
	return &cycledRandomRange{
		stream: stream,
	}
}

func (c *cycledRandomRange) Next() int {
	if val, ok := <-c.stream; ok {
		return val
	}
	return 0
}

type Rand interface {
	Next() int
}

func RangeRandomStream(ctx context.Context, generator Rand) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case out <- generator.Next():
			}
		}
	}()
	return out
}
