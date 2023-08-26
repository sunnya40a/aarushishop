package middleware

import (
	"log"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

const (
	redirectURL = "/login"
	sessionKey  = "user"
)

// AuthRequired is a middleware that checks if the user is authenticated.
// If the user is not logged in, it redirects them to the login page.
func AuthRequired(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get(sessionKey)
	if user == nil {
		log.Println("User not logged in")
		c.Redirect(http.StatusMovedPermanently, redirectURL)
		c.Abort()
		return
	}
	c.Next()
}
