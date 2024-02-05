// * middleware package.
package middleware

import (
	"log"
	"net/http"

	"aarushishop/globals"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

const (
	SessionMaxAge = 10 * 60 // 10 minutes in seconds
)

// AuthMiddlewareAPI handles session expiration and renewal.
func AuthMiddlewareAPI() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if the user is authenticated
		session := sessions.Default(c)
		user := session.Get(globals.UserKey)

		if user == nil {
			// User is not authenticated or session expired
			c.JSON(Unauthorized, gin.H{"content": "Unauthorized request"})
			c.Abort()
			return
		}

		log.Printf("User: %v", user)
		log.Printf("Session middleware activated for user: %v", user)

		// If the session is not expired, renew it by 15 minutes
		session.Options(sessions.Options{
			Path:     "/",
			MaxAge:   SessionMaxAge,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
		})

		// Save the updated session
		if err := session.Save(); err != nil {
			log.Println("Error renewing session:", err)
			c.String(InternalServerError, "Error renewing session")
			c.Abort()
			return
		}

		c.Next()
	}
}
