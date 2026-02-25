package models

import "time"

type Role struct {
	ID          int       `db:"id"`
	RoleName    string    `db:"role_name"`
	Description string    `db:"description"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}
