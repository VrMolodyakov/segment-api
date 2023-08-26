package model

import "time"

type Participation struct {
	UserID      int64
	SegmentName string
	ExpiredAt   time.Time
}
