package models

import "time"

type Client struct {
	ID           []byte    `json:"id" db:"id"`
	ClientName   string    `json:"clientName" db:"client_name"`
	ClientSecret string    `json:"-" db:"client_secret"`
	UpdatedAt    time.Time `json:"updatedAt" db:"updated_at"`
}
