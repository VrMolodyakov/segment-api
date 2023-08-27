package history

import (
	"time"
)

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

type Date struct {
	Year  int
	Month int
}

func (d Date) Validate() error {
	year, month, _ := time.Now().Date()
	if d.Year < AvitoLaunchYear {
		return ErrIncorrectDate
	} else if d.Year == year && month < time.Month(d.Month) {
		return ErrIncorrectDate
	}
	return nil
}
