package models

import (
	"database/sql"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AuthorizationCode struct {
	Code        string       `json:"code" db:"code"`
	ClientId    []byte       `json:"clientId" db:"client_id"`
	UserId      []byte       `json:"userId" db:"user_id"`
	ExpiresAt   time.Time    `json:"expiresAt" db:"expires_at"`
	UsedAt      sql.NullTime `json:"usedAt" db:"used_at"`
	RedirectURI string       `json:"redirectUri" db:"redirect_uri"`
}

type RefreshToken struct {
	ID        int       `json:"id" db:"id"`
	Token     string    `json:"token" db:"token"`
	ClientId  []byte    `json:"clientId" db:"client_id"`
	UserId    []byte    `json:"userId" db:"user_id"`
	ExpiresAt time.Time `json:"expiresAt" db:"expires_at"`
	Revoked   bool      `json:"revoked" db:"revoked"`
}

type UserClaims struct {
	jwt.RegisteredClaims
	UserID     []byte   `json:"userId"`
	Username   string   `json:"userName"`
	FirstName  string   `json:"firstName"`
	MiddleName string   `json:"middleName"`
	LastName   string   `json:"lastName"`
	Email      string   `json:"email"`
	Roles      []string `json:"roles"`
	Scopes     []string `json:"scopes"`
}
