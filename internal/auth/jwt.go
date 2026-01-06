package auth

import (
	"crypto/rsa" // Required for the *rsa.PrivateKey type
	"encoding/hex"
	"fmt"
	"time"

	"github.com/causon-mikolorenz/unified-access-backend/models"
	"github.com/golang-jwt/jwt/v5"
)

// GenerateToken creates a signed OIDC JWT using RS256.
// It accepts the pre-loaded privateKey object for maximum performance.
func GenerateToken(privateKey *rsa.PrivateKey,
	clientID []byte, claims models.UserClaims,
) (string, error) {
	now := time.Now()

	// 1. Convert the binary UUID to a standard hex string for the 'aud' claim.
	clientIDStr := hex.EncodeToString(clientID)

	// 2. Set the standard OIDC registered claims
	claims.RegisteredClaims = jwt.RegisteredClaims{
		Subject:   claims.ID, // Assuming this is the User's unique string ID
		Issuer:    "unified-access-idp",
		Audience:  jwt.ClaimStrings{clientIDStr},
		ExpiresAt: jwt.NewNumericDate(now.Add(1 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		ID:        fmt.Sprintf("%d", now.UnixNano()),
	}

	// 3. Create the token using RS256 (Asymmetric)
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	// 4. Sign the token using the RSA Private Key
	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT: %w", err)
	}

	return signedToken, nil
}
