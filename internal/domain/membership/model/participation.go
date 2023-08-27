package model

import "time"

type MembershipInfo struct {
	UserID      int64
	SegmentName string
	ExpiredAt   time.Time
}

type Membership struct {
	UserID    int64
	SegmentID int64
	ExpiredAt time.Time
}
