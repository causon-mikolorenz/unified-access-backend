package models

import "time"

type ClientGrantType string

const (
	GrantAuthCode          ClientGrantType = "authorization_code"
	GrantRefreshToken      ClientGrantType = "refresh_token"
	GrantClientCredentials ClientGrantType = "client_credentials"
)

func (g ClientGrantType) IsValid() bool {
	switch g {
	case GrantAuthCode, GrantRefreshToken, GrantClientCredentials:
		return true
	}
	return false
}

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

type ClientGrantTypes struct {
	ClientID  []byte          `json:"clientId" db:"client_id"`
	GrantType ClientGrantType `json:"grantType" db:"grant_type"`
}
