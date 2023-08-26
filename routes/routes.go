package routes

import (
	"github.com/gin-gonic/gin"

	controllers "aarushishop/controllers"
)

// PublicRoutes registers the public routes.
func PublicRoutes(g *gin.RouterGroup) {
	g.GET("/login", controllers.LoginGetHandler())
	g.POST("/login", controllers.LoginPostHandler())
	g.GET("/", controllers.IndexGetHandler())
}

// PrivateRoutes registers the private routes.
func PrivateRoutes(g *gin.RouterGroup) {
	g.GET("/dashboard", controllers.DashboardGetHandler())
	g.GET("/logout", controllers.LogoutGetHandler())
	g.POST("/logout", controllers.LogoutPostHandler())
}
