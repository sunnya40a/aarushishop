package main

import (
	"aarushishop/database"
	"aarushishop/globals"
	"aarushishop/handler"
	"context"
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/unrolled/secure"
)

//go:embed frontend/*
var staticFiles embed.FS

func main() {

	if err := database.InitDB(); err != nil {
		panic(err)
	}

	router := gin.Default()

	// Enable CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:8000", "http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Use secure middleware to set security headers
	secureMiddleware := secure.New(secure.Options{
		FrameDeny:          true,
		ContentTypeNosniff: true,
		BrowserXssFilter:   true,
		ContentSecurityPolicy: "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline'; " +
			"style-src 'self' 'unsafe-inline'; " +
			"img-src 'self'; " +
			"connect-src 'self'; " +
			"font-src 'self'; " +
			"frame-src 'self'; " +
			"object-src 'self'",
	})

	router.Use(applySecurityHeaders(secureMiddleware))

	// Serve static files
	router.GET("/assets/*filepath", func(c *gin.Context) {
		filepath := "frontend/assets" + c.Param("filepath")
		content, err := staticFiles.ReadFile(filepath)
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}

		// Set the Content-Type header based on the file extension
		contentType := getContentType(filepath)
		c.Data(http.StatusOK, contentType, content)
	})

	// Handle wildcard route for Vue.js history mode
	router.NoRoute(func(c *gin.Context) {
		indexPath := "frontend/index.html"
		content, err := staticFiles.ReadFile(indexPath)

		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}

		c.Data(http.StatusOK, "text/html; charset=utf-8", content)
	})

	// Serve font files
	router.GET("/fonts/*filepath", func(c *gin.Context) {
		filepath := "fonts" + c.Param("filepath")
		content, err := staticFiles.ReadFile(filepath)
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}

		// Set the Content-Type header based on the file extension
		contentType := getContentType(filepath)
		c.Data(http.StatusOK, contentType, content)
	})

	store := cookie.NewStore(globals.Secret)
	router.Use(sessions.Sessions("my-session", store))

	// Public routes
	public := router.Group("/")
	handler.PublicRoutes(public)

	defer database.CloseDB()

	// Start the server
	server := &http.Server{
		Addr:    "0.0.0.0:8000",
		Handler: router,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

	sig := <-signalChannel
	fmt.Printf("Received signal: %v\n", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	database.CloseDB()

	log.Println("Application gracefully terminated")
}

func applySecurityHeaders(secureMiddleware *secure.Secure) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Apply secure middleware
		err := secureMiddleware.Process(c.Writer, c.Request)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		// Generate a nonce (replace this with a real nonce generator)
		nonce := "your_generated_nonce"

		// Update CSP header to use the nonce for inline scripts
		c.Header("Content-Security-Policy", fmt.Sprintf("default-src 'self'; script-src 'self' 'nonce-%s'; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; img-src 'self'; connect-src 'self'; font-src 'self' https://fonts.gstatic.com; frame-src 'self'; object-src 'self'", nonce))

		// Continue processing the request
		c.Next()
	}
}

func getContentType(filename string) string {
	switch {
	case strings.HasSuffix(filename, ".js"):
		return "application/javascript"
	case strings.HasSuffix(filename, ".css"):
		return "text/css"
	case strings.HasSuffix(filename, ".png"):
		return "image/png"
	case strings.HasSuffix(filename, ".jpg"), strings.HasSuffix(filename, ".jpeg"):
		return "image/jpeg"
	case strings.HasSuffix(filename, ".gif"):
		return "image/gif"
	case strings.HasSuffix(filename, ".svg"):
		return "image/svg+xml"
	// Add more cases for other file types as needed
	default:
		return "application/octet-stream"
	}
}
