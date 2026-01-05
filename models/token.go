package models

import "time"

type AuthorizationCode struct {
	Code      string    `json:"code" db:"code"`
	ClientId  []byte    `json:"clientId" db:"client_id"`
	UserId    []byte    `json:"userId" db:"user_id"`
	ExpiresAt time.Time `json:"expiresAt" db:"expires_at"`
	Used      bool      `json:"used" db:"used"`
}

type RefreshToken struct {
	ID        int       `json:"id" db:"id"`
	Token     string    `json:"token" db:"token"`
	ClientId  []byte    `json:"clientId" db:"client_id"`
	UserId    []byte    `json:"userId" db:"user_id"`
	ExpiresAt time.Time `json:"expiresAt" db:"expires_at"`
	Revoked   bool      `json:"revoked" db:"revoked"`
}
