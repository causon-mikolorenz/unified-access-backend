package models

import "time"

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

type User struct {
	ID           []byte     `json:"id" db:"id"`
	Username     string     `json:"userName" db:"username"`
	FirstName    string     `json:"firstName" db:"first_name"`
	MiddleName   string     `json:"middleName" db:"middle_name"`
	LastName     string     `json:"lastName" db:"last_name"`
	Email        string     `json:"email" db:"email"`
	PasswordHash string     `json:"-" db:"password_hash"`
	Status       UserStatus `json:"status" db:"status"`
	UpdatedAt    time.Time  `json:"updatedAt" db:"updated_at"`

	Roles []string `json:"roles" db:"-"`
}

type Role struct {
	ID          int    `json:"id" db:"id"`
	RoleName    string `json:"roleName" db:"role_name"`
	Description string `json:"description" db:"description"`
}
