package middleware

import (
	"log"
	"net/http"

	"github.com/causon-mikolorenz/unified-access-backend/internal/repository"
	"github.com/gin-gonic/gin"
)

type Middleware struct {
	ClientRepo *repository.ClientRepository
}

/**
 * CORSMiddleware handles cross-origin resource sharing for registered clients.
 * It dynamically fetches allowed base URLs from the client repository.
 */
func (m *Middleware) CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		if origin == "" {
			c.Next()
			return
		}

		baseURLs, err := m.ClientRepo.ListClientBaseURLS()
		if err != nil {
			log.Printf("[CORSMiddleware] Failed to fetch urls: %v", err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		allowedOrigins := make(map[string]bool)
		for _, url := range baseURLs {
			allowedOrigins[url] = true
		}

		if allowedOrigins[origin] {
			h := c.Writer.Header()
			h.Set("Access-Control-Allow-Origin", origin)
			h.Set("Access-Control-Allow-Credentials", "true")
			h.Set("Access-Control-Allow-Methods",
				"GET, POST, PUT, PATCH, DELETE, OPTIONS")
			h.Set("Access-Control-Allow-Headers",
				"Content-Type, Authorization")
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}