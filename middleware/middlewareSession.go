// * middleware package.
package middleware

import (
	"aarushishop/globals"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

const (
	SessionMaxAge = 15 * 60 // 10 minutes in seconds
)

// AuthMiddlewareAPI handles session expiration and renewal.
func AuthMiddlewareAPI() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if the user is authenticated
		session := sessions.Default(c)
		user := session.Get(globals.UserKey)
		storedUserAgent := session.Get("user_agent")
		storedIpAddress := session.Get("ip_address")
		ipAddress := strings.Split(c.Request.RemoteAddr, ":")[0]

		if user == nil {
			// User is not authenticated or session expired
			c.JSON(http.StatusUnauthorized, gin.H{"content": "Unauthorized request"})
			c.Abort()
			return
		}
		// Check if session has expired
		lastAccessTime, ok := session.Get("last_access").(int64) // Retrieve Unix timestamp
		if !ok || time.Since(time.Unix(lastAccessTime, 0)) > (15*time.Minute) {
			c.JSON(http.StatusUnauthorized, gin.H{"status": "session expired"})
			c.Abort()
			return
		}
		//log.Printf("\nLast access time: %s\n", time.Unix(lastAccessTime, 0).Format("2006-01-02 15:04:05"))
		if c.Request.UserAgent()[:min(len(c.Request.UserAgent()), 100)] != storedUserAgent || ipAddress != storedIpAddress {
			// Log potential hijacking attempt and consider invalidating session
			c.JSON(http.StatusUnauthorized, gin.H{"status": "Invalid user enviroment"})
			c.Abort()
			return
		}
		// Check CSRF token validity
		csrfToken := c.GetHeader("X-CSRF-Token")
		if csrfToken == "" || csrfToken != session.Get("X-CSRF-Token") {
			c.JSON(http.StatusUnauthorized, gin.H{"status": "invalid CSRF token"})
			c.Abort()
			return
		}
		// If the session is not expired, renew it
		session.Options(sessions.Options{
			Path:     "/",
			MaxAge:   SessionMaxAge,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
		})
		session.Delete("last_access")
		session.Set("last_access", time.Now().Unix())
		// Save the updated session
		if err := session.Save(); err != nil {
			log.Println("Error renewing session:", err)
			c.String(http.StatusInternalServerError, "Error renewing session")
			c.Abort()
			return
		}
		c.Next()
	}
}
