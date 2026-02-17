package models

import (
	"fmt"
	"time"
)

type UserStatus string

const (
	StatusActive    UserStatus = "active"
	StatusInactive  UserStatus = "inactive"
	StatusSuspended UserStatus = "suspended"
	StatusDeleted   UserStatus = "deleted"
)

func (s UserStatus) CanLogin() bool {
	return s == StatusActive
}

func (s UserStatus) RequiresAction() bool {
	return s == StatusInactive
}

func (s UserStatus) IsRestricted() bool {
	return s == StatusSuspended || s == StatusDeleted
}

func (s UserStatus) MapStatus(status string) (UserStatus, error) {
	switch status {
	case string(StatusActive):
		s = StatusActive
	case string(StatusInactive):
		s = StatusInactive
	case string(StatusSuspended):
		s = StatusSuspended
	case string(StatusDeleted):
		s = StatusDeleted
	default:
		return "", fmt.Errorf("Invalid status string: %s", status)
	}

	return s, nil
}

type User struct {
	ID           []byte     `db:"id"`
	Username     string     `db:"username"`
	FirstName    string     `db:"first_name"`
	MiddleName   string     `db:"middle_name"`
	LastName     string     `db:"last_name"`
	Email        string     `db:"email"`
	PasswordHash string     `db:"password_hash"`
	Status       UserStatus `db:"status"`
	CreatedAt    time.Time  `db:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at"`

	Roles []string `db:"-"`
}

type Role struct {
	ID          int       `db:"id"`
	RoleName    string    `db:"role_name"`
	Description string    `db:"description"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}
