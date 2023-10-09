//* package main.go

package main

import (
	"aarushishop/database"
	"aarushishop/globals"
	"aarushishop/helpers"
	"aarushishop/middleware"
	"context"
	"os/signal"
	"syscall"
	"time"

	//"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	//"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
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
}

func main() {

	if err := database.InitDBPool(); err != nil {
		// Handle the error if the pool initialization fails
		// You might want to log the error and exit the application gracefully
		// or take other appropriate actions
		panic(err)
	}
	// Create a Gin router instance
	router := gin.Default()

	// Serve static files from the "assets" directory
	router.Static("/assets", "./assets")
	router.Static("/favicon.ico", "./assets/favicon.ico")

	// Load HTML templates from the "templates" directory
	templateFiles := []string{}
	err := filepath.Walk("templates", func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".tmpl") {
			templateFiles = append(templateFiles, path)
		}
		return nil
	})
	if err != nil {
		fmt.Println("Error loading templates:", err)
		return
	}

	router.LoadHTMLFiles(templateFiles...)

	// Print the number of loaded templates
	fmt.Println("\nNumber of loaded templates:", len(templateFiles))

	// Print the names of the loaded templates
	fmt.Println("\nLoaded templates:")
	for i, file := range templateFiles {
		fmt.Printf("%d. %s\n", i+1, file)
	}
	//Print empty line.
	fmt.Printf("\n")

	store := cookie.NewStore(globals.Secret)
	router.Use(sessions.Sessions("my-session", store))

	// Create a route group for public routes (not requiring authentication)
	public := router.Group("/")
	PublicRoutes(public)

	// Create a route group for private routes (requiring authentication)
	private := router.Group("/")
	PrivateRoutes(private)

	// Close the database connection pool when your application exits
	defer database.CloseDBPool()
	// Start the server and listen on port 8080
	//router.Run("0.0.0.0:8080")
	//router.RunTLS(":8080", "./cert/localhost.crt", "./cert/localhost.key")
	// Start the server in a separate goroutine
	server := &http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: router,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Capture shutdown signals (Ctrl+C and SIGTERM)
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

	// Block until a shutdown signal is received
	sig := <-signalChannel
	fmt.Printf("Received signal: %v\n", sig)

	// Create a context with a timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Gracefully shut down the server
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	// Close database connections and perform cleanup here
	database.CloseDBPool()

	// Optionally, add more cleanup actions as needed

	log.Println("Application gracefully terminated")
}

//
// LoginPostHandler handles the POST request for user login

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
		log.Println("\nLog in Get event activated")
		//		session := sessions.Default(c)
		//		user := session.Get(globals.UserKey)
		//		if user != nil {
		//			log.Printf("Authorized user tried to access the login page again: %v", user)
		//			c.Redirect(http.StatusFound, "/logout") // Redirect to the logout page
		//			return
		//		}
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
			"content": "This is a dashboard",
			"user":    user,
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
	c.Redirect(http.StatusMovedPermanently, "/login")
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
