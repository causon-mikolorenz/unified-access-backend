package auth

import (
	"fmt"
	"time"

	"github.com/causon-mikolorenz/unified-access-backend/models"
	"github.com/golang-jwt/jwt/v5"
)

func GenerateToken(secretKey string, clientID string, claims models.UserClaims) (string, error) {
	now := time.Now()

	// 1. Set the standard OIDC registered claims
	claims.RegisteredClaims = jwt.RegisteredClaims{
		Subject:   claims.ID,
		Issuer:    "unified-access-idp",
		Audience:  jwt.ClaimStrings{clientID}, // Identify which app this token is for
		ExpiresAt: jwt.NewNumericDate(now.Add(1 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),           // Token is valid immediately
		ID:        fmt.Sprintf("%d", now.UnixNano()), // Unique JTI
	}

	// 2. Create and sign the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT: %w", err)
	}

	return signedToken, nil
}
