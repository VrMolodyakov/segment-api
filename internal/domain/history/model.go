package history

import (
	"fmt"
	"strconv"
	"time"
)

const (
	AvitoLaunchYear int    = 2007
	timeFormat      string = "2006-01-02 15:04:05"
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

func (d Date) ToString() string {
	return fmt.Sprintf("%d-%d", d.Year, d.Month)
}

func NewDate(year int, month int) Date {
	return Date{
		Year:  year,
		Month: month,
	}
}

func (d Date) Validate() error {
	year, month, _ := time.Now().Date()
	if d.Year < AvitoLaunchYear {
		return ErrIncorrectYear
	} else if d.Year == year && month < time.Month(d.Month) {
		return ErrIncorrectMonth
	}
	return nil
}

func (h History) Row() []string {
	return []string{
		strconv.FormatInt(h.ID, 10),
		strconv.FormatInt(h.UserID, 10),
		h.Segment,
		string(h.Operation),
		h.Time.Format(timeFormat),
	}
}

func (h History) Headers() []string {
	return []string{"ID", "UserID", "Segment", "Operation", "Time"}
}
