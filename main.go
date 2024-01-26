package main

import (
	"aarushishop/database"
	"aarushishop/globals"
	"aarushishop/handler"
	"aarushishop/middleware"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func main() {
	if err := database.InitDB(); err != nil {
		panic(err)
	}

	router := gin.Default()

	// Enable CORS

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"}, // Replace with the correct origin of your Vue.js app
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	//router.Use(middleware.CSPMiddleware()) // It block cdn and java script outside the origin.
	router.Use(middleware.STSMiddleware())
	router.Use(middleware.SecurityHeadersMiddleware())

	// Serve static files
	router.Static("/assets", "./assets")
	router.Static("/static", "./static") // Corrected directory path
	router.Static("/favicon.ico", "./assets/favicon.ico")

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

	fmt.Println("\nNumber of loaded templates:", len(templateFiles))

	fmt.Println("\nLoaded templates:")
	for i, file := range templateFiles {
		fmt.Printf("%d. %s\n", i+1, file)
	}
	fmt.Printf("\n")

	store := cookie.NewStore(globals.Secret)
	router.Use(sessions.Sessions("my-session", store))

	public := router.Group("/")
	handler.PublicRoutes(public)

	private := router.Group("/")
	handler.PrivateRoutes(private)

	defer database.CloseDB()

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
