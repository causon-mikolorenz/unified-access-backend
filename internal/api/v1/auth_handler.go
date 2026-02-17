package v1

import (
	"crypto/rsa"
	"log"
	"net/http"
	"time"

	"github.com/causon-mikolorenz/unified-access-backend/internal/auth"
	"github.com/causon-mikolorenz/unified-access-backend/internal/dto"
	"github.com/causon-mikolorenz/unified-access-backend/internal/repository"
	"github.com/causon-mikolorenz/unified-access-backend/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthHandler struct {
	Repo        *repository.AuthCodeRepository
	SessionRepo *repository.SessionRepository
	PrivateKey  *rsa.PrivateKey
}

// LoginAndAuthorize verifies credentials and issues an authorization code
// @Summary Login and Authorize
// @Description Authenticate user and return a redirect URL with auth code
// @Tags Authentication
// @Accept json
// @Produce json
// @Param login body dto.LoginRequest true "Login Credentials"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) LoginAndAuthorize(c *gin.Context) {
	var req dto.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[LoginAndAuthorize] What failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	claims, storedHash, err := h.Repo.GetUserForAuth(req.Email)
	if err != nil {
		log.Printf("[LoginAndAuthorize] What failed: email %s not found", req.Email)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	if err := auth.CompareSecret(storedHash, req.Password); err != nil {
		log.Printf("[LoginAndAuthorize] What failed: invalid pass for %s", req.Email)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	code, err := auth.GenerateAuthorizationCode()
	if err != nil {
		log.Printf("[LoginAndAuthorize] What failed: code generation %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	clientUUID, err := uuid.Parse(req.ClientID)
	if err != nil {
		log.Printf("[LoginAndAuthorize] What failed: client_id parse %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid client_id"})
		return
	}

	var clientIDBytes [16]byte
	copy(clientIDBytes[:], clientUUID[:])

	registeredURI, err := h.Repo.GetClientRedirectURI(clientIDBytes[:])
	if err != nil {
		log.Printf("[LoginAndAuthorize] What failed: client lookup %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_client"})
		return
	}

	if req.RedirectURI != registeredURI {
		log.Printf("[LoginAndAuthorize] What failed: uri mismatch %s", req.Email)
		c.JSON(http.StatusForbidden, gin.H{"error": "unauthorized_redirect_uri"})
		return
	}

	err = h.Repo.StoreCode(code, claims.UserID, clientIDBytes[:], req.RedirectURI)
	if err != nil {
		log.Printf("[LoginAndAuthorize] What failed: store code %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	sessionID, _ := auth.GenerateRandomString(32)
	expiry := time.Now().AddDate(SESSION_YEARS, SESSION_MONTHS, SESSION_DAYS)

	newSession := &models.IdPSession{
		SessionId: sessionID,
		UserId:    claims.UserID,
		IpAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		ExpiresAt: expiry,
	}

	if err := h.SessionRepo.Create(newSession); err != nil {
		log.Printf("[LoginAndAuthorize] What failed: session create %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "session error"})
		return
	}

	maxAge := int(time.Hour.Seconds() * 24 * SESSION_DAYS)
	c.SetCookie("idp_session", sessionID, maxAge, "/", "", true, true)

	c.JSON(http.StatusOK, gin.H{
		"redirect_to": req.RedirectURI + "?code=" + code,
	})
}

// Logout terminates the user session and revokes all tokens
// @Summary Global Logout
// @Description Clear session cookie and revoke all issued tokens for the user
// @Tags Authentication
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	sessionID, err := c.Cookie("idp_session")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "already logged out"})
		return
	}

	session, err := h.SessionRepo.GetByID(sessionID)
	if err != nil {
		c.SetCookie("idp_session", "", -1, "/", "", true, true)
		c.JSON(http.StatusOK, gin.H{"message": "session cleared"})
		return
	}

	if err := h.Repo.RevokeTokens(session.UserId); err != nil {
		log.Printf("[Logout] What failed: revocation %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "logout failed"})
		return
	}

	c.SetCookie("idp_session", "", -1, "/", "", true, true)
	c.JSON(http.StatusOK, gin.H{"message": "global logout successful"})
}

// CheckSession verifies if the current session cookie is valid
// @Summary Check Session
// @Description Validate the idp_session cookie against the database
// @Tags Authentication
// @Produce json
// @Success 200 {object} map[string]bool
// @Failure 401 {object} map[string]bool
// @Router /api/v1/auth/session [get]
func (h *AuthHandler) CheckSession(c *gin.Context) {
	sessionID, err := c.Cookie("idp_session")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"authenticated": false})
		return
	}

	session, err := h.SessionRepo.GetByID(sessionID)
	if err != nil || time.Now().After(session.ExpiresAt) {
		c.JSON(http.StatusUnauthorized, gin.H{"authenticated": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"authenticated": true,
		"user_id":       session.UserId,
	})
}