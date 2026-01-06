package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

func GenerateAuthorizationCode() (string, error) {
	// 32 bytes of entropy provides 256 bits of security,
	// which is well above the standard requirements for OAuth2.
	b := make([]byte, 32)

	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("failed to generate secure bytes: %w", err)
	}

	// Use RawURLEncoding to avoid padding characters (=) which
	// can sometimes cause issues in URL query strings.
	return base64.RawURLEncoding.EncodeToString(b), nil
}
