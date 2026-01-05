package models

import "time"

type User struct {
	ID           []byte    `json:"id" db:"id"`
	Username     string    `json:"userName" db:"username"`
	FirstName    string    `json:"firstName" db:"first_name"`
	MiddleName   string    `json:"middleName" db:"middle_name"`
	LastName     string    `json:"lastName" db:"last_name"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	Status       string    `json:"status" db:"status"`
	UpdatedAt    time.Time `json:"updatedAt" db:"updated_at"`

	Roles []string `json:"roles" db:"-"`
}
