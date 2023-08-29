package membership

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

type FullMembershipInfo struct {
	UserID      int64
	SegmentID   int64
	SegmentName string
	ExpiredAt   time.Time
}
