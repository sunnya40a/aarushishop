// helpers/helpers.go
package helpers

import (
	"aarushishop/database"
	"context"
	"log"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// CheckUserPass checks if the given username and password are valid
func CheckUserPass(username, password string) bool {
	// Check if the username and password are empty
	if EmptyUserPass(username, password) {
		return false
	}

	// Get the database connection
	db := database.GetDB() // Use the GetDB function for MariaDB

	// Query the database to find the user's hashed password
	var hashedPassword string
	err := db.QueryRowContext(context.Background(), "SELECT password_hash FROM users WHERE username = ?", username).Scan(&hashedPassword)
	if err != nil {
		log.Println("Error querying the database:", err)
		return false
	}

	// Compare the hashed password with the provided password using bcrypt
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		log.Println("Password verification failed:", err)
		return false
	}

	log.Println("User Password OK")
	return true
}

// GeneratePasswordHash generates a password hash using bcrypt with a specified cost factor.
func GeneratePasswordHash(password string, cost int) (string, error) {
	// Generate a salted hash for the password with the specified cost factor
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", err
	}

	return string(hashedPassword), nil
}

// EmptyUserPass checks if the given username and password are empty
func EmptyUserPass(username, password string) bool {
	return strings.TrimSpace(username) == "" || strings.TrimSpace(password) == ""
}
