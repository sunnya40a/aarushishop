//* handler.go

package handler

import (
	"github.com/gin-gonic/gin"
)

func PublicRoutes(g *gin.RouterGroup) {
	// Public routes (do not require authentication)
	g.POST("loginapi", LoginAPI())
	g.POST("logoutapi", LogoutAPI())

	purchase := g.Group("/purchase")
	{
		purchase.POST("/add", AddPurchaseAPI())
		purchase.GET("/list", ListPurchaseAPI())
		purchase.POST("/adduser", AddTestUserAPI())
	}
}
