package v1

import (
	"bytes"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/causon-mikolorenz/unified-access-backend/internal/auth"
	"github.com/gin-gonic/gin"
)

func (h *AuthHandler) ExchangeToken(c *gin.Context) {
	// 1. Parse Request
	code := c.PostForm("code")
	clientIDString := c.PostForm("client_id")
	clientSecret := c.PostForm("client_secret")

	// Convert the incoming client_id string to binary for DB verification
	// We strip dashes to handle standard UUID strings correctly
	cleanHex := strings.ReplaceAll(clientIDString, "-", "")
	clientIDBin, err := hex.DecodeString(cleanHex)
	if err != nil || len(clientIDBin) != 16 {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": "invalid_client_id_format"},
		)
		return
	}

	// 2. Verify Client Identity (using binary ID and Bcrypt secret)
	validClient, err := h.Repo.VerifyClient(clientIDBin, clientSecret)
	if err != nil || !validClient {
		c.JSON(http.StatusUnauthorized,
			gin.H{"error": "unauthorized_client"},
		)
		return
	}

	// 3. Consume the Code (Atomic transaction handles the 'used_at' check)
	authCode, err := h.Repo.ExchangeCode(code)
	if err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": "invalid_grant", "message": err.Error()},
		)
		return
	}

	// 4. Security Check: Ensure this client owns the requested code
	if !bytes.Equal(authCode.ClientId, clientIDBin) {
		c.JSON(http.StatusForbidden,
			gin.H{"error": "invalid_grant", "message": "client mismatch"},
		)
		return
	}

	// 5. Fetch User Profile/Claims using the UserID stored in the code
	claims, err := h.Repo.GetClaimsByID(authCode.UserId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error"})
		return
	}

	// 6. Generate the JWT (RS256)
	// We pass the pre-loaded h.PrivateKey from our handler struct
	accessToken, err := auth.GenerateToken(h.PrivateKey, clientIDBin, *claims)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": "failed_to_sign_token"},
		)
		return
	}

	// 7. Return OIDC Compliant Response
	c.JSON(http.StatusOK, gin.H{
		"access_token": accessToken,
		"token_type":   "Bearer",
		"expires_in":   3600, // 1 hour expiration
	})
}
