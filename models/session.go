package models

import "time"

type IdPSession struct {
	SessionId string    `json:"sessionId" db:"session_id"`
	UserId    []byte    `json:"userId" db:"user_id"`
	IpAddress string    `json:"ipAddress" db:"ip_address"`
	UserAgent string    `json:"userAgent" db:"user_agent"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	ExpiresAt time.Time `json:"expiresAt" db:"expires_at"`
}
