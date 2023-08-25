package model

import "time"

type Operation string

var (
	Deleted = Operation("deleted")
	Added   = Operation("added")
)

type History struct {
	ID        int64
	UserID    int64
	Segment   string
	Operation Operation
	Time      time.Time
}
