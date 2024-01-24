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

	g.GET("/test", TestAPI())
	g.GET("/login", LoginGetHandler())
	g.POST("/login", LoginPostHandler())
	g.GET("/", IndexGetHandler())
	g.POST("/loginapi",LoginAPI())
	api := g.Group("/api")
	{
		api.GET("/myuser", ListMyUserAPI())        
		//learning Purpose only
		api.GET("/users", ListUserAPI())                // Get a list of users.
		api.GET("/users/:client_id", GetuserbyIDAPI())  // Get a user by ID.
		api.POST("/users", CreateUserAPI())             // Create a new user.
		api.PUT("/users/:client_id", EditUserAPI())     // Modify a user by ID.
		api.DELETE("/users/:client_id",DeleteUserAPI()) // Delete a user by ID.

		api.GET("/category",ListCategoryAPI())
	}
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
	
	// below are for experiment.
	g.GET("/learn", LearnTableGetHandler()) // Learning purpose only.
	g.GET("/entry", LearnEntryGetHandler()) // Learning purpose only.

	// api := g.Group("/api")
	// {
	// 	api.GET("/myuser", ListMyUserAPI())        
	// 	//learning Purpose only
	// 	api.GET("/users", ListUserAPI())                // Get a list of users.
	// 	api.GET("/users/:client_id", GetuserbyIDAPI())  // Get a user by ID.
	// 	api.POST("/users", CreateUserAPI())             // Create a new user.
	// 	api.PUT("/users/:client_id", EditUserAPI())     // Modify a user by ID.
	// 	api.DELETE("/users/:client_id",DeleteUserAPI()) // Delete a user by ID.
	// }
}
