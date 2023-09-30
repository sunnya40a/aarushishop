package controllers

import (
	"log"
	"net/http"

	globals "aarushishop/globals"
	helpers "aarushishop/helpers"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

const (
	LoginTemplate     = "login.tmpl"
	DashboardTemplate = "dashboard.tmpl"
)

// Initialize a secure cookie store
var store = cookie.NewStore([]byte(globals.Secret))

func init() {
	// Set session options for security
	secureCookie := true // Set to true if you're using HTTPS
	httpOnlyCookie := true

	store.Options(sessions.Options{
		Secure:   secureCookie,   // Set to true if you're using HTTPS
		HttpOnly: httpOnlyCookie, // Set to true to make cookies accessible only via HTTP
		SameSite: http.SameSiteStrictMode,
	})
}

// LoginGetHandler handles the GET request for the login page
func LoginGetHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("Log in Get event activated")
		session := sessions.Default(c)
		user := session.Get(globals.Userkey)
		if user != nil {
			log.Printf("Authorized user tried to access the login page again: %v", user)
			c.Redirect(http.StatusFound, "/logout") // Redirect to the logout page
			return
		}
		c.HTML(http.StatusOK, LoginTemplate, gin.H{
			"content": "",
			"user":    user,
		})
	}
}

// LoginPostHandler handles the POST request for user login
func LoginPostHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("Log in post event activated")
		username := c.PostForm("username")
		password := c.PostForm("password")

		if helpers.EmptyUserPass(username, password) {
			c.HTML(http.StatusBadRequest, LoginTemplate, gin.H{"content": "Parameters can't be empty."})
			return
		}

		// Check user credentials without using ctx
		if !helpers.CheckUserPass(username, password) {
			c.HTML(http.StatusUnauthorized, LoginTemplate, gin.H{"content": "Incorrect username or password."})
			return
		}

		session := sessions.Default(c)
		session.Set(globals.Userkey, username)

		// Set the session expiration time (15 minutes in this example)
		session.Options(sessions.Options{
			MaxAge: 5 * 60, // 5 minutes in seconds
		})

		if err := session.Save(); err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.Redirect(http.StatusMovedPermanently, "/dashboard")
	}
}

// LogoutPostHandler handles the POST request for user logout
func LogoutPostHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("Logout Post event activated")
		logoutUser(c)
	}
}

// LogoutGetHandler handles the GET request for user logout
func LogoutGetHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("Logout Get event activated")
		logoutUser(c)
	}
}

// IndexGetHandler handles the GET request for the index page
func IndexGetHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("IndexGetHandler activated")
		c.HTML(http.StatusOK, "index.tmpl", gin.H{})
	}
}

// DashboardGetHandler handles the GET request for the dashboard page
func DashboardGetHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("DashboardGetHandler activated")
		session := sessions.Default(c)
		user := session.Get(globals.Userkey)

		if user == nil {
			c.HTML(http.StatusUnauthorized, LoginTemplate, gin.H{"content": "User not found in session."})
			return
		}

		c.HTML(http.StatusOK, DashboardTemplate, gin.H{
			"content": "This is a dashboard",
			"user":    user,
		})
	}
}

// logoutUser deletes the user from the session and saves it
func logoutUser(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get(globals.Userkey)
	log.Println("Logging out user:", user)

	if user == nil {
		log.Println("Invalid session token.")
		c.HTML(http.StatusMovedPermanently, LoginTemplate, gin.H{"content": "Invalid session token."})
		return
	}

	session.Delete(globals.Userkey)
	log.Println("Session delete activated")
	if err := session.Save(); err != nil {
		log.Println("Failed to save session:", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Redirect(http.StatusMovedPermanently, "/login")
}
