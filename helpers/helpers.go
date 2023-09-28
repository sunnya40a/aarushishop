package helpers

import (
	// Import the ConnectToDB function
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

	// Get a database connection
	db, err := database.ConnectToDB()
	if err != nil {
		log.Println("Error connecting to the database:", err)
		return false
	}
	defer db.Close(context.Background())

	// Query the database to find the user's hashed password
	var hashedPassword string
	err = db.QueryRow(context.Background(), "SELECT password_hash FROM users WHERE username = $1", username).Scan(&hashedPassword)
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

	return true
}

// EmptyUserPass checks if the given username and password are empty
func EmptyUserPass(username, password string) bool {
	return strings.Trim(username, " ") == "" || strings.Trim(password, " ") == ""
}
