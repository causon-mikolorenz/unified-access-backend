package auth

import (
	"encoding/hex" // Required for binary to hex string conversion
	"fmt"
	"time"

	"github.com/causon-mikolorenz/unified-access-backend/models"
	"github.com/golang-jwt/jwt/v5"
)

// GenerateToken creates a signed OIDC JWT.
// We accept clientID as []byte to match your repository, but convert to string for the JWT.
func GenerateToken(secretKey string, clientID []byte, claims models.UserClaims) (string, error) {
	now := time.Now()

	// 1. Convert the binary UUID to a standard hex string.
	// This ensures the "aud" (Audience) claim is readable by the Service Provider.
	clientIDStr := hex.EncodeToString(clientID)

	// 2. Set the standard OIDC registered claims
	claims.RegisteredClaims = jwt.RegisteredClaims{
		Subject: claims.ID,
		Issuer:  "unified-access-idp",
		// Audience expects []string. We wrap our hex string in the slice.
		Audience:  jwt.ClaimStrings{clientIDStr},
		ExpiresAt: jwt.NewNumericDate(now.Add(1 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		ID:        fmt.Sprintf("%d", now.UnixNano()),
	}

	// 3. Create and sign the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT: %w", err)
	}

	return signedToken, nil
}
