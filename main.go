//* package main.go

package main

import (
	"aarushishop/database"
	"aarushishop/globals"
	"aarushishop/handler"

	//"aarushishop/model"
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
	router.Static("/static", ".static")
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
	handler.PublicRoutes(public)

	// Create a route group for private routes (requiring authentication)
	private := router.Group("/")
	handler.PrivateRoutes(private)

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