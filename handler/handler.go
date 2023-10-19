//* handler.go

package handler

import (
	"aarushishop/middleware"

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
	g.GET("/dashboard", DashboardGetHandler())  // Display the user's dashboard.
	g.GET("/logout", LogoutGetHandler())  // Handle user logout (GET).
	g.POST("/logout", LogoutGetHandler())  // Handle user logout (POST).
	// Additional required private  route here.
	g.GET("/table", TableGetHandler())  // Display a data table.
	g.GET("/listuser", ListUserGetHandler())  // Listing users

	// Create a sub-group for API routes
	api := g.Group("/api")
	{
		api.GET("/listusers", APIListUserHandler())  // Get a list of users.
		// Add more API routes here.
	}
}
