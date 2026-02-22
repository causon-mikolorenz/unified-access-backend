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
	ID            []byte    `db:"id"`
	ClientName    string    `db:"client_name"`
	Abbreviation  string    `db:"abbreviation"`
	ClientSecret  string    `db:"client_secret"`
	BaseUrl       string    `db:"base_url"`
	RedirectUri   string    `db:"redirect_uri"`
	LogoutUri     string    `db:"logout_uri"`
	Description   string    `db:"description"`
	ImageLocation string    `db:"image_location"`
	UpdatedAt     time.Time `db:"updated_at"`
}

type ClientGrantTypes struct {
	ClientID  []byte          `json:"clientId" db:"client_id"`
	GrantType ClientGrantType `json:"grantType" db:"grant_type"`
}
