package models

import (
	"time"
)

type User struct {
	ID        int64
	CreatedAt time.Time
	UpdatedAt time.Time
	Name      string
	Age       *int64
	Surname   *string
	Email     string
	Password  string
}
