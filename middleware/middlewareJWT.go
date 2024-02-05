// * middleware package.
package middleware

import (
	"log"
	"strings"

	"aarushishop/globals"
	"aarushishop/model"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Middleware to validate JWT token
func AuthMiddlewareAPIJWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(Unauthorized, gin.H{"error": "Missing Authorization header"})
			c.Abort()
			return
		}
		// Extract token string from Bearer authorization header
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.ParseWithClaims(tokenString, &model.CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
			return globals.JwtSecretKey, nil
		})

		if err != nil {
			c.JSON(Unauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(*model.CustomClaims)
		if !ok || !token.Valid {
			c.JSON(Unauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		log.Printf("JWT Middleware activated")
		c.Set("username", claims.Username)
		c.Next()
	}
}
