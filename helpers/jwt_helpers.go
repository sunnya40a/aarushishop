// helpers/jwt_helpers.go
package helpers

import (
	"aarushishop/globals"
	"aarushishop/model"
	"errors"
	"log"

	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)
func CreateRefreshToken(username string) (string, error) {
    claims := &model.RefreshClaims{
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 1)), // Expire in a week
        },
        Username: username,
    }
    refreshTokentoken := jwt.NewWithClaims(jwt.SigningMethodHS384, claims)
    refreshTokenString, err := refreshTokentoken.SignedString(globals.RefreshSecretKey) // Use a different secret for refresh tokens

    if err != nil {
        return "", err
    }
    return refreshTokenString, nil
}

// CreateToken generates a JWT token for the given username using Gin.
func CreateToken(username string) (string, error) {
	claims := &model.CustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 10)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS384, claims)
	tokenString, err := token.SignedString(globals.JwtSecretKey)

	if err != nil {
		return "", err
	}
	return tokenString, nil
}
func RefreshTokenAPI() gin.HandlerFunc {
    return func(c *gin.Context) {
        refreshTokenString := c.GetHeader("refreshtoken")
        if refreshTokenString == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing RefreshToken header"})
            return
        }

		// Extract token string from Bearer authorization header
		refreshTokenString = strings.TrimPrefix(refreshTokenString, "Bearer ")

        // Validate the refresh token (use a function like below)
        claims, err := validateRefreshToken(refreshTokenString)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid RefreshToken"})
            return
        }

        // Issue a new access token
        accessToken, err := CreateToken(claims.Username)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
            return
        }
		log.Printf("\nNew toke issued %s\n",accessToken)
        c.JSON(http.StatusOK, gin.H{"newtoken": "Bearer " + accessToken})
    }
}

func validateRefreshToken(refreshTokenString string) (*model.RefreshClaims, error) {
    token, err := jwt.ParseWithClaims(refreshTokenString, &model.RefreshClaims{}, func(token *jwt.Token) (interface{}, error) {
        return globals.RefreshSecretKey, nil
    })
    if err != nil {
        return nil, err
    }

    // Check token validity and expiration
    if !token.Valid {
        return nil, errors.New("invalid refreshtoken")
    }

    claims, ok := token.Claims.(*model.RefreshClaims)
    if !ok {
        return nil, errors.New("invalid refreshtoken claims")
    }

    return claims, nil
}