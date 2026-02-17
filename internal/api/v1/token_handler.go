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

// PostTokenExchange handles the exchange of an auth code for access tokens
// @Summary Exchange Auth Code
// @Description Validates the code and client secret to issue JWT and Refresh
// @Tags Authentication
// @Accept json
// @Produce json
// @Param req body dto.TokenExchangeRequest true "Exchange Request"
// @Success 200 {object} dto.TokenResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/auth/token [post]
func (h *AuthHandler) PostTokenExchange(c *gin.Context) {
	var req dto.TokenExchangeRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[PostTokenExchange] Bind JSON Error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}

	clientUUID, err := uuid.Parse(req.ClientID)
	if err != nil {
		log.Printf("[PostTokenExchange] Client ID Parse Error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_client"})
		return
	}
	clientIDBin := clientUUID[:]

	valid, err := h.Repo.VerifyClient(clientIDBin, req.ClientSecret)
	if err != nil || !valid {
		log.Printf("[PostTokenExchange] Client Verification Failed: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	authCode, err := h.Repo.ExchangeCode(req.Code)
	if err != nil {
		log.Printf("[PostTokenExchange] Code Exchange Error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_grant"})
		return
	}

	if !bytes.Equal(authCode.ClientId, clientIDBin) {
		log.Printf("[PostTokenExchange] Client ID Mismatch Error")
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid_grant"})
		return
	}

	claims, err := h.Repo.GetClaimsByID(authCode.UserId)
	if err != nil {
		log.Printf("[PostTokenExchange] Claims Retrieval Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error"})
		return
	}

	accessToken, err := auth.GenerateToken(h.PrivateKey, clientIDBin, *claims)
	refreshStr, err := auth.GenerateRandomString(64)
	_ = h.Repo.StoreRefreshToken(refreshStr, claims.UserID, clientIDBin)

	c.JSON(http.StatusOK, dto.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshStr,
		ExpiresIn:    3600,
		TokenType:    "Bearer",
	})
}

// PostTokenRotate handles refreshing an access token using a refresh token
// @Summary Rotate Refresh Token
// @Description Invalidates old refresh token and issues a new pair
// @Tags Authentication
// @Accept json
// @Produce json
// @Param req body dto.RefreshRequest true "Refresh Request"
// @Success 200 {object} dto.TokenResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) PostTokenRotate(c *gin.Context) {
	var req dto.RefreshRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[PostTokenRotate] Bind JSON Error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}

	uID, cID, err := h.Repo.GetIDsFromToken(req.RefreshToken)
	if err != nil {
		log.Printf("[PostTokenRotate] Token Lookup Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid_token"})
		return
	}

	newToken, err := auth.GenerateRandomString(64)
	if err != nil {
		log.Printf("[PostTokenRotate] Token Generation Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token_error"})
		return
	}

	if err := h.Repo.RotateRefreshToken(req.RefreshToken, newToken); err != nil {
		log.Printf("[PostTokenRotate] Token Rotation Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "rotate_fail"})
		return
	}

	claims, err := h.Repo.GetClaimsByID(uID)
	if err != nil {
		log.Printf("[PostTokenRotate] Claims Retrieval Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error"})
		return
	}

	accessToken, _ := auth.GenerateToken(h.PrivateKey, cID, *claims)

	c.JSON(http.StatusOK, dto.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newToken,
		ExpiresIn:    3600,
		TokenType:    "Bearer",
	})
}