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

const (
	SESSION_YEARS  = 0
	SESSION_MONTHS = 0
	SESSION_DAYS   = 15
)

func (h *AuthHandler) LoginAndAuthorize(c *gin.Context) {
	var req dto.LoginRequest

	// 1. Bind JSON from React
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// 2. Fetch user (using req.Email instead of c.PostForm)
	claims, storedHash, err := h.Repo.GetUserForAuth(req.Email)
	if err != nil {
		log.Printf("[Login] Invalid email. Email: %s", req.Email)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// 3. Verify password (using req.Password)
	if err := auth.CompareSecret(storedHash, req.Password); err != nil {
		log.Printf("[Login] Invalid password for email %s", req.Email)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// 4. Generate code
	code, err := auth.GenerateAuthorizationCode()
	if err != nil {
		log.Printf("[Login] Failed Authorization Code Generation: %w", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate code"})
		return
	}

	// 5. Store Code
	clientUUID, err := uuid.Parse(req.ClientID)
	if err != nil {
		log.Printf("[Login] Parsing failed for client_id: %w", err)
		c.JSON(http.StatusBadRequest,
			gin.H{"error": "invalid client_id format"},
		)
		return
	}

	// Convert UUID to the exact 16-byte array format
	var clientIDBytes [16]byte
	copy(clientIDBytes[:], clientUUID[:])

	registeredURI, err := h.Repo.GetClientRedirectURI(clientIDBytes[:])
	if err != nil {
		log.Printf("[Login] Client lookup failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_client"})
		return
	}

	if req.RedirectURI != registeredURI {
		log.Printf("[Login] Redirect URI mismatch. Expected: %s, Got: %s",
			registeredURI, req.RedirectURI,
		)
		c.JSON(http.StatusForbidden,
			gin.H{"error": "unauthorized_redirect_uri"},
		)
		return
	}

	// Pass clientIDBytes[:] to the repo
	err = h.Repo.StoreCode(code, claims.UserID, clientIDBytes[:], req.RedirectURI)
	if err != nil {
		log.Printf("[Login] StoreCode Database Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	sessionID, _ := auth.GenerateRandomString(32)

	newSession := &models.IdPSession{
		SessionId: sessionID,
		UserId:    claims.UserID,
		IpAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		ExpiresAt: time.Now().AddDate(SESSION_YEARS, SESSION_MONTHS, SESSION_DAYS),
	}

	if err := h.SessionRepo.Create(newSession); err != nil {
		log.Printf("[Login] Session Creation Error: %v", err)
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": "session_creation_failed"},
		)
		return
	}

	maxAge := int(time.Hour.Seconds() * 24 * SESSION_DAYS)
	c.SetCookie("idp_session", sessionID, maxAge, "/", "", true, true)

	redirectURL := req.RedirectURI + "?code=" + code
	c.JSON(http.StatusOK, gin.H{
		"redirect_to": redirectURL,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
    sessionID, err := c.Cookie("idp_session")
    if err != nil {
        c.JSON(http.StatusOK, gin.H{"message": "already logged out"})
        return
    }

    session, err := h.SessionRepo.GetByID(sessionID)
    if err != nil {
        // Cookie exists but DB doesn't recognize it. Clear it and exit.
        c.SetCookie("idp_session", "", -1, "/", "", true, true)
        c.JSON(http.StatusOK, gin.H{"message": "session cleared"})
        return
    }

    // Call the Global Logout Procedure
    if err := h.Repo.RevokeTokens(session.UserId); err != nil {
        log.Printf("[Logout] Global Revocation Failed: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "logout failed"})
        return
    }

    // Clear the browser's entry point
    c.SetCookie("idp_session", "", -1, "/", "", true, true)
    c.JSON(http.StatusOK, gin.H{"message": "global logout successful"})
}

func (h *AuthHandler) CheckSession(c *gin.Context) {
	sessionID, err := c.Cookie("idp_session")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"authenticated": false})
		return
	}

	// Verify session still exists and is not expired in DB
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
