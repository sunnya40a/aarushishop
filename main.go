package main

import (
	globals "aarushishop/globals"
	middleware "aarushishop/middleware"
	routes "aarushishop/routes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func main() {
	// Create a Gin router instance
	router := gin.Default()

	// Serve static files from the "assets" directory
	router.Static("/assets", "./assets")

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

	// Create a cookie store for session management, using the secret key from globals
	store := cookie.NewStore([]byte(globals.Secret))

	// Use the sessions middleware with the "session" name and the cookie store
	router.Use(sessions.Sessions("session", store))

	// Create a route group for public routes (not requiring authentication)
	public := router.Group("/")
	routes.PublicRoutes(public)

	// Create a route group for private routes (requiring authentication)
	private := router.Group("/")
	private.Use(middleware.AuthRequired) // Apply the AuthRequired middleware to all routes in this group
	routes.PrivateRoutes(private)

	// Start the server and listen on port 8080
	router.Run("0.0.0.0:8080")
}
