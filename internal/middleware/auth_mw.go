package middleware

import (
	"crypto/rsa"
	"fmt"
	"net/http"
	"strings"

	"github.com/causon-mikolorenz/unified-access-backend/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthorizeRBAC validates an RS256 JWT and checks for required roles.
// It takes the pre-loaded publicKey from main.go to avoid disk I/O on every request.
func AuthorizeRBAC(publicKey *rsa.PublicKey,
	requiredRoles ...string,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Extract the Authorization Header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized,
				gin.H{"error": "Authorization header is required"},
			)
			c.Abort()
			return
		}

		// 2. Validate Bearer format
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized,
				gin.H{"error": "Authorization header must be Bearer token"},
			)
			c.Abort()
			return
		}

		// 3. Parse and Validate the Token using the Public Key
		claims := &models.UserClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims,
			func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v",
						token.Header["alg"],
					)
				}
				// Return the pre-loaded public key
				return publicKey, nil
			})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized,
				gin.H{"error": "Invalid or expired token"},
			)
			c.Abort()
			return
		}

		// 4. Role-Based Access Control (RBAC) Check
		if len(requiredRoles) > 0 {
			authorized := false
			for _, userRole := range claims.Roles {
				for _, reqRole := range requiredRoles {
					if userRole == reqRole {
						authorized = true
						break
					}
				}
				if authorized {
					break
				}
			}

			if !authorized {
				c.JSON(http.StatusForbidden,
					gin.H{"error": "Insufficient permissions"},
				)
				c.Abort()
				return
			}
		}

		// 5. Context Injection
		c.Set("user_id", claims.UserID)
		c.Set("user_claims", claims)

		c.Next()
	}
}
