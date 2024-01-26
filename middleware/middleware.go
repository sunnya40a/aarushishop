// * middleware package.
package middleware

import (
	"aarushishop/globals"
	"log"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// Strict Transport Security (STS) configuration
func STSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		//Strict-Transport-Security: max-age=31536000; includeSubDomains; preload
		c.Next()
	}
}

// Content Security Policy (CSP)
// It block cdn and java script outside the origin.
func CSPMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		//c.Writer.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' https://trusted-cdn.com;")
		c.Writer.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self';")
		//Content-Security-Policy: default-src 'self'; script-src 'self' https://trusted-cdn.com
		c.Next()
	}
}

// Middleware for Headers:
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("X-Content-Type-Options", "nosniff")
		c.Writer.Header().Set("X-Frame-Options", "DENY")
		//c.Writer.Header().Set("Referrer-Policy", "no-referrer")
		c.Writer.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		//c.Writer.Header().Set("Content-Security-Policy", "default-src 'self' localhost; script-src 'self' https://xychhabi34.com;")
		c.Writer.Header().Set("Content-Security-Policy", "default-src 'self' *; script-src 'self' 'unsafe-inline' 'unsafe-eval' *;")
		c.Writer.Header().Set("X-Permitted-Cross-Domain-Policies", "none")
		c.Writer.Header().Set("Feature-Policy", "geolocation 'self'; microphone 'self'; camera 'self'")
		// Add more headers as needed

		c.Next()
	}
}

// SessionMiddleware handles session expiration and renewal.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if the user is authenticated
		session := sessions.Default(c)
		user := session.Get(globals.UserKey)

		if user == nil {
			// User is not authenticated or session expired
			//c.Redirect(http.StatusSeeOther, "/login")
			c.HTML(http.StatusSeeOther, "login.tmpl", gin.H{
				"content": "You are not authorized. Please log in",
			})
			c.Abort()
			return
		}
		log.Print("\n============\nSession middleware activated\n")
		// If the session is not expired, renew it by 15 minutes
		session.Options(sessions.Options{
			MaxAge:   60 * 15, // 15 Min
			SameSite: http.SameSiteStrictMode,
			Secure:   true,
			HttpOnly: true,
		})

		// Save the updated session
		if err := session.Save(); err != nil {
			c.String(http.StatusInternalServerError, "Error renewing session")
			c.Abort()
			return
		}

		c.Next()
	}
}
