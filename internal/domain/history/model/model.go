package model

import "time"

type Operation string

var (
	Deleted = Operation("deleted")
	Added   = Operation("added")
)

type Histpry struct {
	ID        int
	UserID    int
	Segment   string
	Operation Operation
	Time      time.Time
}
