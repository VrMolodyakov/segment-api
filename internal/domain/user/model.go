package user

import (
	"regexp"
)

var (
	emailRegex = regexp.MustCompile(`^.+@[^\.].*\.[a-z]{2,}$`)
)

type UserID int

type User struct {
	ID        int64
	FirstName string
	LastName  string
	Email     string
}

func (u User) Valid() error {
	if !emailRegex.MatchString(u.Email) {
		return ErrInvalidEmail
	}
	return nil
}
