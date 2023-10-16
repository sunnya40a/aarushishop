//* handler.go

package handler

import (
	"aarushishop/globals"
	"aarushishop/helpers"
	"aarushishop/middleware"

	//"context"
	"log"
	"net/http"

	//"time"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

const (
	LoginTemplate     = "login.tmpl"
	DashboardTemplate = "dashboard.tmpl"
	TableTemplate     = "table.tmpl"
)

func PublicRoutes(g *gin.RouterGroup) {
	// Public routes (do not require authentication)
	g.GET("/login", LoginGetHandler())
	g.POST("/login", LoginPostHandler())
	g.GET("/", IndexGetHandler())
}

func PrivateRoutes(g *gin.RouterGroup) {
	// Apply the AuthMiddleware to protect these routes
	g.Use(middleware.AuthMiddleware())

	// Define your private routes here
	g.GET("/dashboard", DashboardGetHandler())
	g.GET("/table", TableGetHandler())
	g.GET("/logout", LogoutGetHandler())
	g.POST("/logout", LogoutGetHandler())
	g.GET("/testing", TestingGetHandler())
}

func LoginPostHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("\nLog in post event activated")
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
		log.Println("\nIndexGetHandler activated")
		c.HTML(http.StatusOK, "index.tmpl", gin.H{})
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

	// Close the database connection (replace with your database close logic)
	// Assuming you have a function CloseDB() to close the database connection
	// Replace this with your actual database close logic

	// Redirect to the login page after successful logout
	c.Redirect(http.StatusMovedPermanently, "/")
	log.Print("Logout Successful.")
}

func TableGetHandler() gin.HandlerFunc {
	log.Println("\nTableGet Handler activated")
	return func(c *gin.Context) {
		session := sessions.Default(c)
		user := session.Get(globals.UserKey)

		if user == nil {
			c.HTML(http.StatusUnauthorized, LoginTemplate, gin.H{"content": "User not found in session."})
			return
		}

		// If the user is found in the session, render the HTML page with user data.
		c.HTML(http.StatusOK, "table.tmpl", gin.H{
			//"content": "This is a dashboard",
			"user": user,
		})
	}
}

func TestingGetHandler() gin.HandlerFunc {
	log.Println("\nTesting Handler activated")
	return func(c *gin.Context) {
		session := sessions.Default(c)
		user := session.Get(globals.UserKey)

		if user == nil {
			c.HTML(http.StatusUnauthorized, LoginTemplate, gin.H{"content": "User not found in session."})
			return
		}

		// If the user is found in the session, render the HTML page with user data.
		c.HTML(http.StatusOK, "vue.tmpl", gin.H{
			//"content": "This is a dashboard",
		// You can pass any data that your Vue.js template may need here.
		// For example, if you need to pass the user data:			
			"user": user,
		})
	}
}

/* func TestingGetHandler() gin.HandlerFunc {
	log.Println("\nTesting Handler activated")
	return func(c *gin.Context) {
		session := sessions.Default(c)
		user := session.Get(globals.UserKey)

		if user == nil {
			c.HTML(http.StatusUnauthorized, LoginTemplate, gin.H{"content": "User not found in session."})
			return
		}

		// Connect to the database (assuming you've set up the DB connection)
		dbConn, err := database.GetDBConnection()
		if err != nil {
			c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"message": "Database connection error"})
			return
		}
		defer dbConn.Release()

		// Execute the SQL query to fetch data from the "users" table
		rows, err := dbConn.Query(context.Background(), "SELECT user_id, username, email, password_hash, comment FROM users")
		if err != nil {
			c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"message": "Failed to fetch data from the database"})
			return
		}
		defer rows.Close()

		// Create a slice to store the user data
		var users []model.User // Replace "model.User" with the struct type that matches your user data

		// Iterate through the query results and append them to the slice
		for rows.Next() {
			model.user
			var user model.User // Replace "model.User" with the struct type that matches your user data
			if err := rows.Scan(&user.UserID, &user.Username, &user.Email, &user.PasswordHash, &user.Comment); err != nil {
				c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"message": "Error scanning user data"})
				return
			}
			users = append(users, user)
		}

		c.HTML(http.StatusOK, "vue.tmpl", gin.H{
			"users": users,
		})
	}
} */