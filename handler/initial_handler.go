package handler

import (
	"aarushishop/globals"
	"aarushishop/helpers"
	"aarushishop/model"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/mrz1836/go-sanitize"
)

const (
	SessionMaxAge = 15 * 60 // 10 minutes in seconds
)

// setSessionOptions sets up session options for the user.
func setSessionOptions(c *gin.Context) {
	session := sessions.Default(c)
	session.Options(sessions.Options{
		Path:     "/",
		MaxAge:   SessionMaxAge,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
}

// LoginAPI handles user login.
func LoginAPI() gin.HandlerFunc {
	return func(c *gin.Context) {
		var validtoken bool
		var user model.LoginUser
		var tokenString, refreshToken string

		// Step 1: Validate and bind JSON data from the request to the model.LoginUser struct
		if err := c.ShouldBindJSON(&user); err != nil {
			log.Println("Error on JSON Binding:", err.Error())
			c.JSON(BadRequest, gin.H{"content": "Invalid JSON format"})
			return
		}

		// Sanitize user input before using it in HTML responses
		user.Username = sanitize.HTML(sanitize.Scripts(helpers.SanitizeUsername(user.Username)))
		user.Password = helpers.SanitizePassword(user.Password)

		//below 3 line is to print tokenString. we can remove it latter.
		authHeader1 := c.GetHeader("authorization")
		tokenString1 := strings.TrimPrefix(authHeader1, "Bearer ")
		log.Printf("Token Verified:%s", tokenString1)

		// Step 2: Check if username and password are empty
		if helpers.EmptyUserPass(user.Username, user.Password) {
			c.JSON(BadRequest, gin.H{"content": "Parameters can't be empty."})
			return
		}

		// Step 3: Check if the provided username and password are valid
		if !helpers.CheckUserPass(user.Username, user.Password) {
			c.JSON(Unauthorized, gin.H{"content": "Incorrect username or password."})
			return
		}

		// Step 4: Set up session options and store the user's username in the session
		setSessionOptions(c)
		session := sessions.Default(c)
		csrfToken, err := helpers.GenerateCSRFToken()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate CSRF token"})
			return
		}
		session.Set("X-CSRF-Token", csrfToken)
		session.Set("last_access", time.Now().Unix())
		session.Set("user_agent", c.Request.UserAgent()[:min(len(c.Request.UserAgent()), 100)])
		session.Set("ip_address", strings.Split(c.Request.RemoteAddr, ":")[0])

		// Store signed user data in the session
		userData := map[string]string{
			"user": user.Username,
		}
		signedData, err := helpers.SignData(userData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to sign session data"})
			return
		}
		session.Set(globals.UserKey, signedData)
		log.Printf("User %s Passed from session", user.Username)
		// Step 5: Retrieve the JWT token from the Authorization header
		authHeader := c.GetHeader("authorization")
		if authHeader == "" {
			validtoken = false
		} else {
			// Extract token string from Bearer authorization header
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")

			// Step 6: Parse and validate the JWT token
			_, err := parseAndValidateJWT(tokenString)
			if err != nil {
				validtoken = false
			} else {
				validtoken = true
			}
		}

		// Step 7: For this example, create and return a new JWT token and refresh token if no valid token is present
		if !validtoken {
			if err := handleJWTTokenCreation(c, user.Username, &tokenString, &refreshToken); err != nil {
				return
			}
		}
		c.Header("X-CSRF-Token", csrfToken)
		// Save the session
		if err := session.Save(); err != nil {
			log.Println("Error saving session:", err)
			c.JSON(InternalServerError, gin.H{"content": "Internal Server Error"})
			return
		}
		// Step 8: Respond with appropriate JSON based on the presence of a valid token
		if validtoken {
			log.Printf("User %s logged in successfully with existing token", user.Username)
			c.JSON(OK, gin.H{
				"xcsrftoken": csrfToken,
				"content":    "Login Successful with existing token",
			})
		} else {
			log.Printf("User %s logged in successfully with new token", user.Username)
			c.JSON(OK, gin.H{
				"token":      "Bearer " + tokenString,
				"reftoken":   "Bearer " + refreshToken,
				"xcsrftoken": csrfToken,
				"content":    "Login successful...",
			})
		}
	}
}

// Function parseAndValidateJWT parses and validates the JWT token.
func parseAndValidateJWT(tokenString string) (*jwt.Token, error) {
	token, err := jwt.ParseWithClaims(tokenString, &model.CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return globals.JwtSecretKey, nil
	})

	if err != nil {
		//Token is invalid
		return nil, fmt.Errorf("invalid token")
	}

	// Check if the token is valid
	_, ok := token.Claims.(*model.CustomClaims)
	if !ok || !token.Valid {
		//Token is invalid
		return nil, fmt.Errorf("invalid token")
	}
	return token, nil
}

// Function handleJWTTokenCreation handles the creation of JWT token and refresh token.
func handleJWTTokenCreation(c *gin.Context, username string, tokenString, refreshToken *string) error {
	var err error
	*tokenString, err = helpers.CreateToken(username)
	if err != nil {
		log.Println("Error creating JWT token:", err)
		c.JSON(InternalServerError, gin.H{"content": "Internal Server Error"})
		return fmt.Errorf("error creating JWT token: %v", err)
	}

	*refreshToken, err = helpers.CreateRefreshToken(username)
	if err != nil {
		log.Println("Error creating Refresh token:", err)
		c.JSON(InternalServerError, gin.H{"content": "Internal Server Error"})
		return fmt.Errorf("error creating Refresh token: %v", err)
	}

	return nil
}

// LogoutAPI handles user logout.
func LogoutAPI() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Retrieve the user from the session
		session := sessions.Default(c)
		//user := session.Get(globals.UserKey)
		///
		signedData := session.Get(globals.UserKey)
		if signedData == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"status": "invalid session data"})
			c.Abort()
			return
		}

		userData, err := helpers.VerifyData(signedData.(string))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"status": "invalid session data"})
			c.Abort()
			return
		}
		user := userData[globals.UserKey]
		if user == "" {
			// Handle the case where the session is invalid or the user is not logged in
			c.JSON(http.StatusUnauthorized, gin.H{"content": "invalid session data."})
			return
		}
		// Delete the user from the session
		session.Options(sessions.Options{
			MaxAge:   0, // Invalidate the session immediately
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
		})

		// Clear user information from session
		session.Clear()
		// Clear signed user data
		session.Delete(globals.UserKey)
		session.Delete("last_access")
		session.Delete("user_agent")
		session.Delete("ip_address")
		// Clear CSRF token
		session.Delete("X-CSRF-Token")

		// Save the session to remove the user's session cookie
		if err := session.Save(); err != nil {
			log.Println("Error saving session during logout:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"content": "Logout failed. Please try again."})
			return
		}
		log.Printf("User %s logged out successfully", user)
		c.JSON(OK, gin.H{"content": "Logout successful..."})
	}
}

// LogoutAPI handles user logout using token but I have to do nothing in serverside as client delete its token on logout.
// This function is here as we can modify in future.
func LogoutAPIJWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Retrieve the JWT token from the Authorization header
		tokenString := c.GetHeader("authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"content": "Missing Authorization header"})
			return
		}
		// Extract token string from Bearer authorization header
		tokenString = strings.TrimPrefix(tokenString, "Bearer ")
		// Parse and validate the JWT token
		token, err := jwt.ParseWithClaims(tokenString, &model.CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
			return globals.JwtSecretKey, nil
		})
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"content": "Invalid token"})
			return
		}
		// Check if the token is valid
		if !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"content": "Invalid token"})
			return
		}
		// Extract user information from the token
		claims, ok := token.Claims.(*model.CustomClaims)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"content": "Failed to extract claims"})
			return
		}
		// Now you have the user information, and you can proceed with the logout process if needed.
		// In this example, we'll just log the user out.
		log.Printf("User %s logged out successfully", claims.Username)
		c.JSON(http.StatusOK, gin.H{"content": "Logout successful..."})
	}
}
