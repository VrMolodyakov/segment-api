package cache

import "time"

type cleaner struct {
	interval time.Duration
	stop     chan struct{}
}

func (c *cleaner) stopCleaner() {
	c.stop <- struct{}{}
}
