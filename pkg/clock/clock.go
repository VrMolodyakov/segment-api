package clock

import (
	"time"
)

type Clock interface {
	Now() time.Time
	Since(t time.Time) time.Duration
	Until(t time.Time) time.Duration
}

type clock struct{}

func New() Clock {
	return &clock{}
}

func (c *clock) Now() time.Time { return time.Now() }

func (c *clock) Since(t time.Time) time.Duration { return time.Since(t) }

func (c *clock) Until(t time.Time) time.Duration { return time.Until(t) }
