package model

type UserID int

type User struct {
	ID        int64
	FirstName string
	LastName  string
	Email     string
}
