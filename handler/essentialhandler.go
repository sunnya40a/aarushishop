//* essentialhandler.go

package handler

import (
	"aarushishop/globals"
	"aarushishop/helpers"
	"aarushishop/model"
	"log"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)


func LoginPostHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.PostForm("Username")
		password := c.PostForm("Password")

		if helpers.EmptyUserPass(username, password) {
			c.HTML(http.StatusBadRequest, LoginTemplate, gin.H{"content": "Parameters can't be empty."})
			return
		}

		// Check user credentials without using ctx
		if !helpers.CheckUserPass(username, password) {
			c.HTML(http.StatusUnauthorized, LoginTemplate, gin.H{"content": "Incorrect username or password."})
			return
		}

		// Create a session for the authenticated user with custom options
		session := sessions.Default(c)
		session.Options(sessions.Options{
			Path:     "/",
			MaxAge:   900, // 15 minutes in seconds
			HttpOnly: true,
			Secure:   true, // Set to true if your application uses HTTPS
			SameSite: http.SameSiteStrictMode,
		})
		session.Set(globals.UserKey, username)

		// Save the session (set the session cookie)
		if err := session.Save(); err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.Redirect(http.StatusSeeOther, "/dashboard")
	}
}

// LoginGetHandler handles the GET request for the login page
func LoginGetHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		user := session.Get(globals.UserKey)

		// You can check if the user is already authenticated and redirect them to the dashboard
		// or simply show the login page.
		if user != nil {
			c.HTML(http.StatusOK, DashboardTemplate, gin.H{
				"content": "Your session is still valid",
				"user":    user,
			})
			return
		}
		// If the user is not authenticated, show the login page with a "200 OK" status code.
		c.HTML(http.StatusOK, LoginTemplate, gin.H{})
	}
}

// LogoutPostHandler handles the POST request for user logout
func LogoutPostHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		logoutUser(c)
	}
}

// LogoutGetHandler handles the GET request for user logout
func LogoutGetHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		logoutUser(c)
	}
}

// IndexGetHandler handles the GET request for the index page
func IndexGetHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "indextable.tmpl", gin.H{})
	}
}

// DashboardGetHandler handles the GET request for the dashboard page
func DashboardGetHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		user := session.Get(globals.UserKey)

		if user == nil {
			c.HTML(http.StatusUnauthorized, LoginTemplate, gin.H{"content": "User not found in session."})
			return
		}

		c.HTML(http.StatusOK, DashboardTemplate, gin.H{
			//"content": "This is a dashboard",
			"user": user,
		})
	}
}

// logoutUser deletes the user session and performs cleanup actions
func logoutUser(c *gin.Context) {
	// Retrieve the user from the session
	session := sessions.Default(c)
	user := session.Get(globals.UserKey)

	if user == nil {
		// Handle the case where the session is invalid or user is not logged in
		c.HTML(http.StatusMovedPermanently, LoginTemplate, gin.H{"content": "Invalid session token."})
		return
	}

	// Delete the user from the session
	session.Delete(globals.UserKey)

	// Save the session to remove the user's session cookie
	if err := session.Save(); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// Redirect to the login page after successful logout
	c.Redirect(http.StatusMovedPermanently, "/")
	log.Print("\n========================\nSession Deleted and Logout Successful.\n")
}

func LoginAPI() gin.HandlerFunc {
	return func(c *gin.Context) {

		var user model.LoginUser
		if err := c.ShouldBindJSON(&user); err != nil {
			log.Printf("Error on JSON Binding: %v", err.Error())
            c.JSON(http.StatusBadRequest, gin.H{"content": "Invalid JSON format"})
            return
        }

		// Check if username or password is empty
		log.Printf("Username :  %s -- Password: %s", user.Username, user.Password)
		if helpers.EmptyUserPass(user.Username, user.Password) {
			c.JSON(http.StatusBadRequest, gin.H{"content": "Parameters can't be empty."})
			return
		}
		log.Printf("%s %s",user.Username, user.Password)
		// Check user credentials
		if !helpers.CheckUserPass(user.Username, user.Password) {
			// Use constant for status code
			c.JSON(http.StatusUnauthorized, gin.H{"content": "Incorrect username or password."})
			return
		}

		// Create a session for the authenticated user with custom options
		session := sessions.Default(c)
		session.Options(sessions.Options{
			Path:     "/",
			MaxAge:   900, // 15 minutes in seconds
			HttpOnly: true,
			Secure:   true, // Set to true if your application uses HTTPS
			SameSite: http.SameSiteStrictMode,
		})

		// Set the authenticated user in the session
		session.Set(globals.UserKey, user.Username)

		// Save the session (set the session cookie)
		if err := session.Save(); err != nil {
			// Log the error for debugging purposes
			log.Printf("Error saving session: %v", err)
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}


		// Log successful login
		log.Printf("User %s logged in successfully", user.Username)

		// Optionally, you may send a success response
		c.JSON(http.StatusOK, gin.H{"content": "Login successful..."})
	}
}

func TestAPI() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200,gin.H{
			"message": "Axios is working nicely",
		})
}
}