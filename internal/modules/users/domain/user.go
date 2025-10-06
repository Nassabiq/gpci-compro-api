package domain

import "time"

type User struct {
	ID       int64
	XID      string
	Name     string
	Email    string
	Password string
	IsActive bool

	CreatedAt *time.Time
	UpdatedAt *time.Time
}
