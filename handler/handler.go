//* handler.go

package handler

import (
	"aarushishop/middleware"

	"github.com/gin-gonic/gin"
)

func PublicRoutes(g *gin.RouterGroup) {
	// Public routes (do not require authentication)
	g.POST("loginapi", LoginAPI())
	g.POST("logoutapi", LogoutAPI())
	//private route
	purchase := g.Group("/purchase")
	purchase.Use(middleware.AuthMiddlewareAPI())
	{
		purchase.POST("/add", AddPurchaseAPI())
		purchase.GET("/list", ListPurchaseAPI())
	}
}
