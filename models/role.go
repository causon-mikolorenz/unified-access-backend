package models

type Role struct {
	ID          int    `json:"id" db:"id"`
	RoleName    string `json:"roleName" db:"role_name"`
	Description string `json:"description" db:"description"`
}
