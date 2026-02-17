package v1

import (
	"crypto/rsa"
	"net/http"

	"github.com/causon-mikolorenz/unified-access-backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

type Handlers struct {
	AuthHandler   *AuthHandler
	ClientHandler *ClientHandler
	RoleHandler   *RoleHandler
	UserHandler   *UserHandler
	PubKey        *rsa.PublicKey
	CORS          gin.HandlerFunc
}

func MapRoutes(v1Group *gin.RouterGroup, h Handlers) {
	auth := v1Group.Group("/auth")
	{
		auth.POST("/login", h.AuthHandler.LoginAndAuthorize)
		auth.POST("/token", h.AuthHandler.PostTokenExchange)
		auth.POST("/refresh", h.AuthHandler.PostTokenRotate)
		auth.POST("/logout", h.AuthHandler.Logout)
		auth.GET("/session", h.AuthHandler.CheckSession)
	}

	// Protected Admin Endpoints
	admin := v1Group.Group("/admin")
	admin.Use(middleware.AuthorizeRBAC(h.PubKey, "idp:admin"))
	{
		admin.GET("/status", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "IdP is operational"})
		})

		// Service Provider (Client) Maintenance
		clients := admin.Group("/clients")
		{
			clients.POST("", h.ClientHandler.PostClient)
			clients.GET("", h.ClientHandler.GetClientList)
			clients.GET("/:id", h.ClientHandler.GetClient)
			clients.PUT("/:id", h.ClientHandler.PutClient)
			clients.DELETE("/:id", h.ClientHandler.DeleteClient)
		}

		// Role Maintenance
		roles := admin.Group("/roles")
		{
			roles.POST("", h.RoleHandler.PostRole)
			roles.GET("", h.RoleHandler.GetRoleList)
			roles.GET("/:id", h.RoleHandler.GetRole)
			roles.PUT("/:id", h.RoleHandler.PutRole)
			roles.DELETE("/:id", h.RoleHandler.DeleteRole)
		}

		// User Maintenance
		users := admin.Group("/users")
		{
			users.POST("", h.UserHandler.PostUser)
			users.GET("", h.UserHandler.GetUserList)
			users.GET("/:id", h.UserHandler.GetUser)
			users.DELETE("/:id", h.UserHandler.DeleteUser)
		}
	}
}
