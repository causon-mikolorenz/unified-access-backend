package v1

import (
	"bytes"
	"log"
	"net/http"

	"github.com/causon-mikolorenz/unified-access-backend/internal/auth"
	"github.com/causon-mikolorenz/unified-access-backend/internal/dto"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *AuthHandler) ExchangeToken(c *gin.Context) {
	var req dto.TokenExchangeRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[ExchangeToken] Bind JSON Error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}

	clientUUID, err := uuid.Parse(req.ClientID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_client"})
		return
	}
	clientIDBin := clientUUID[:]

	validClient, err := h.Repo.VerifyClient(clientIDBin, req.ClientSecret)
	if err != nil || !validClient {
		log.Printf("[ExchangeToken] Client Verification Failed: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized_client"})
		return
	}

	// 4. Consume the Code
	authCode, err := h.Repo.ExchangeCode(req.Code)
	if err != nil {
		// THIS LOG WILL SHOW IF IT'S EXPIRED OR ALREADY USED
		log.Printf("[ExchangeToken] ExchangeCode Error for code [%s]: %v", req.Code, err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_grant",
			"message": err.Error(),
		})
		return
	}

	if !bytes.Equal(authCode.ClientId, clientIDBin) {
		log.Printf("[ExchangeToken] Client ID Mismatch: Code owned by %x, requested by %x", authCode.ClientId, clientIDBin)
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid_grant", "message": "client mismatch"})
		return
	}

	claims, err := h.Repo.GetClaimsByID(authCode.UserId)
	if err != nil {
		log.Printf("[ExchangeToken] GetClaimsByID Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error"})
		return
	}

	accessToken, err := auth.GenerateToken(h.PrivateKey, clientIDBin, *claims)

	refreshTokenStr, err := auth.GenerateRandomString(64)
	err = h.Repo.StoreRefreshToken(refreshTokenStr, claims.UserID, clientIDBin)

	c.JSON(http.StatusOK, dto.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenStr,
		ExpiresIn:    3600,
		TokenType:    "Bearer",
	})
}

func (h *AuthHandler) RotateToken(c *gin.Context) {
	var req dto.RefreshRequest
	
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[RotateToken] Bind JSON Error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}

	userID, clientID, err := h.Repo.GetIDsFromToken(req.RefreshToken)
	if err != nil {
		log.Printf("[RotateToken] Error getting user from token: %w", err)
		c.JSON(http.StatusInternalServerError, 
				gin.H{"error": "user_lookup_failed"},
		)
	}

	newToken, err := auth.GenerateRandomString(64)
	if err != nil {
		log.Printf("[RotateToken] Error generating refresh token: %w", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token_error"})
		return
	}

	err = h.Repo.RotateRefreshToken(req.RefreshToken, newToken)
	if err != nil {
		log.Printf("[RotateToken] Error rotating token: %w", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "rotation_error"})
		return
	}

	claims, err := h.Repo.GetClaimsByID(userID)
	if err != nil {
		log.Printf("[ExchangeToken] GetClaimsByID Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error"})
		return
	}

	accessToken, err := auth.GenerateToken(h.PrivateKey, clientID, *claims)

	c.JSON(http.StatusOK, dto.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newToken,
		ExpiresIn:    3600,
		TokenType:    "Bearer",
	})	
}