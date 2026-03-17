package models

import (
	"time"
)

type User struct {
	ID        int64
	CreatedAt time.Time
	UpdatedAt time.Time
	Data      UserData
}

type UserData struct {
	Name     string
	Age      *int64
	Surname  *string
	Email    string
	Password string
}
