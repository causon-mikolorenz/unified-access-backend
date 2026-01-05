package models

import "time"

type Client struct {
	ID           []byte    `json:"id" db:"id"`
	ClientName   string    `json:"clientName" db:"client_name"`
	ClientSecret string    `json:"-" db:"client_secret"`
	UpdatedAt    time.Time `json:"updatedAt" db:"updated_at"`
}

type ClientUrl struct {
	ID          int    `json:"id" db:"id"`
	ClientID    []byte `json:"clientId" db:"client_id"`
	RedirectURL string `json:"redirectUrl" db:"redirect_url"`
}

type ClientGrantType struct {
	ClientID  []byte `json:"clientId" db:"client_id"`
	GrantType string `json:"grantType" db:"grant_type"`
}
