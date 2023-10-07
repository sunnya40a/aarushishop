// * middleware package.
package middleware

import (
	"aarushishop/globals"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

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
				"content": fmt.Sprintf("User is not authenticated. %s.", user),
			})
			c.Abort()
			return
		}
		log.Print("\nSession middleware activated\n")
		// If the session is not expired, renew it by 15 minutes
		session.Options(sessions.Options{
			MaxAge:   900, // 15 minutes in seconds
			HttpOnly: true,
			Secure:   true, // Set to true if your application uses HTTPS
			SameSite: http.SameSiteStrictMode,
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
