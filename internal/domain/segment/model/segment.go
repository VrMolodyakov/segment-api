package model

import "time"

type Segment struct {
	ID        int64
	Name      string
	ExpiredAt time.Time
}

type SegmentInfo struct {
	ID        int64
	ExpiredAt time.Time
}
