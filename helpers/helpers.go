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
// CheckUserPass checks if the given username and password are valid
func CheckUserPass(username, password string) bool {
	// Check if the username and password are empty
	if EmptyUserPass(username, password) {
		return false
	}

	// Get a database connection from the pool
	db, err := database.GetDBConnection()
	if err != nil {
		log.Println("Error: failed to obtain a database connection to check pass.")
		return false
	}
	defer db.Release()

	// Query the database to find the user's hashed password
	var hashedPassword string
	err = db.QueryRow(context.Background(), "SELECT password_hash FROM users WHERE username = $1", username).Scan(&hashedPassword)
	if err != nil {
		log.Println("Error querying the database:", err)
		return false
	}
	log.Println("Query OK")

	// Compare the hashed password with the provided password using bcrypt
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		log.Println("Password verification failed:", err)
		return false
	}
	log.Println("User Password OK")
	return true
}

// EmptyUserPass checks if the given username and password are empty
func EmptyUserPass(username, password string) bool {
	return strings.TrimSpace(username) == "" || strings.TrimSpace(password) == ""
}
