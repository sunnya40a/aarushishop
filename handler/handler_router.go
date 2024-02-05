//* handler.go

package handler

import (
	"aarushishop/helpers"
	"aarushishop/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	BadRequest          = http.StatusBadRequest
	Unauthorized        = http.StatusUnauthorized
	InternalServerError = http.StatusInternalServerError
	OK                  = http.StatusOK
)

func PublicRoutes(g *gin.RouterGroup) {
	// Public routes (do not require authentication)
	g.POST("loginapi", LoginAPI())
	g.POST("refreshtoken", helpers.RefreshTokenAPI())
}

func PrivateRoutes(g *gin.RouterGroup) {
	// Use session expiration/renewal middleware
	g.Use(middleware.AuthMiddlewareAPI())
	// Use JWT token validation middleware
	g.Use(middleware.AuthMiddlewareAPIJWT())
	// Define private routes
	g.POST("logoutapi", LogoutAPI())
	// Group for purchase-related routes
	purchase := g.Group("/purchase")
	{
		purchase.POST("/add", AddPurchaseAPI())
		purchase.GET("/list", ListPurchaseAPI())
	}
}
