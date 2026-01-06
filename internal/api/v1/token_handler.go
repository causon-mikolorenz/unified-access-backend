package v1

import (
	"bytes"
	"net/http"
	"os"

	"github.com/causon-mikolorenz/unified-access-backend/internal/auth"
	"github.com/gin-gonic/gin"
)

func (h *AuthHandler) ExchangeToken(c *gin.Context) {
	// 1. Parse Request (Standard OAuth2 uses form-urlencoded)
	code := c.PostForm("code")
	clientIDString := c.PostForm("client_id")
	clientSecret := c.PostForm("client_secret")

	clientID := []byte(clientIDString)

	// 2. Verify Client Identity
	validClient, err := h.Repo.VerifyClient(clientID, clientSecret)
	if err != nil || !validClient {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized_client"})
		return
	}

	// 3. Consume the Code (Atomic transaction in your repo)
	authCode, err := h.Repo.ExchangeCode(code)
	if err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": "invalid_grant", "message": err.Error()},
		)
		return
	}

	// 4. Security Check: Client mismatch
	// Ensure the client asking for the token is the same one that requested the code
	if !bytes.Equal(authCode.ClientId, clientID) {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": "invalid_grant", "message": "client mismatch"},
		)
		return
	}

	// 5. Fetch User Claims for the JWT
	claims, _, err := h.Repo.GetClaimsByID(authCode.UserId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error"})
		return
	}

	// 6. Generate the JWT using internal/auth/jwt.go
	accessToken, err := auth.GenerateToken(os.Getenv("JWT_SECRET"),
		clientID, *claims)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error"})
		return
	}

	// 7. Return OIDC Compliant Response
	c.JSON(http.StatusOK, gin.H{
		"access_token": accessToken,
		"token_type":   "Bearer",
		"expires_in":   3600, // 1 hour
	})
}
