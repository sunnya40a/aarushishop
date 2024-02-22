// * middleware package.
package middleware

import (
	"net/http"
)

const (
	InternalServerError = http.StatusInternalServerError
	OK                  = http.StatusOK
	BadRequest          = http.StatusBadRequest
	Unauthorized        = http.StatusUnauthorized
)

// // Strict Transport Security (STS) configuration
// func STSMiddleware() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
// 		//Strict-Transport-Security: max-age=31536000; includeSubDomains; preload
// 		c.Next()
// 	}
// }

// // Middleware for Headers:
// func SecurityHeadersMiddleware() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		c.Writer.Header().Set("X-Content-Type-Options", "nosniff")
// 		c.Writer.Header().Set("X-Frame-Options", "DENY")
// 		//c.Writer.Header().Set("Referrer-Policy", "no-referrer")
// 		c.Writer.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
// 		//c.Writer.Header().Set("Content-Security-Policy", "default-src 'self' localhost; script-src 'self' https://xychhabi34.com;")
// 		c.Writer.Header().Set("Content-Security-Policy", "default-src 'self' *; script-src 'self' 'unsafe-inline' 'unsafe-eval' *;")
// 		c.Writer.Header().Set("X-Permitted-Cross-Domain-Policies", "none")
// 		c.Writer.Header().Set("Feature-Policy", "geolocation 'self'; microphone 'self'; camera 'self'")
// 		// Add more headers as needed

// 		c.Next()
// 	}
// }
