package v1

import (
	"crypto/rsa"
	"net/http"

	"github.com/causon-mikolorenz/unified-access-backend/internal/auth"
	"github.com/causon-mikolorenz/unified-access-backend/repository"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	Repo       *repository.AuthCodeRepository
	PrivateKey *rsa.PrivateKey
}

func (h *AuthHandler) LoginAndAuthorize(c *gin.Context) {
	// 1. Get credentials and OAuth2 params from form
	email := c.PostForm("email")
	password := c.PostForm("password")
	clientID := c.PostForm("client_id")
	redirectURI := c.PostForm("redirect_uri")

	// 2. Fetch user & hash from DB using your Repository
	claims, storedHash, err := h.Repo.GetUserForAuth(email)
	if err != nil {
		c.JSON(http.StatusUnauthorized,
			gin.H{"error": "invalid email or password"},
		)
		return
	}

	// 3. Verify password using your internal/auth/hash.go
	if err := auth.CompareSecret(storedHash, password); err != nil {
		c.JSON(http.StatusUnauthorized,
			gin.H{"error": "invalid email or password"},
		)
		return
	}

	// 4. Generate high-entropy code using internal/auth/codes.go
	code, err := auth.GenerateAuthorizationCode()
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": "failed to generate code"},
		)
		return
	}

	// 5. Save the code to DB using your Repository StoreCode
	// Note: clientID is converted to []byte as per your repo method
	err = h.Repo.StoreCode(code, claims.UserID, []byte(clientID), redirectURI)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": "database error"},
		)
		return
	}

	// 6. Redirect back to the Service Provider with the code
	c.Redirect(http.StatusFound, redirectURI+"?code="+code)
}
