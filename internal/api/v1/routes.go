package v1

import (
	"crypto/rsa"
	"net/http"

	"github.com/causon-mikolorenz/unified-access-backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

type Handlers struct {
	AuthHandler *AuthHandler
	PubKey *rsa.PublicKey
}

func MapRoutes(v1Group *gin.RouterGroup, h Handlers) {
	auth := v1Group.Group("/auth")
	{
		auth.POST("/login", h.AuthHandler.LoginAndAuthorize)
		auth.POST("/token", h.AuthHandler.ExchangeToken)
		auth.POST("/refresh", h.AuthHandler.RotateToken)
		auth.POST("/logout", h.AuthHandler.Logout)

		auth.GET("/session", h.AuthHandler.CheckSession)
	}

	// Protected Admin Endpoints (Using the pre-loaded pubKey)
	admin := v1Group.Group("/admin")
	admin.Use(middleware.AuthorizeRBAC(h.PubKey, "idp:admin"))
	{
		admin.GET("/status", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "IdP is operational"})
		})
	}
}
