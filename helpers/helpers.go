// helpers/helpers.go
package helpers

import (
	"aarushishop/database"
	"aarushishop/globals"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
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

func GenerateCSRFToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

func SignData(data map[string]string) (string, error) {
	// Marshal data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	// Generate HMAC signature using secret key and session salt
	h := hmac.New(sha256.New, []byte(globals.Secret+globals.SessionSalt))
	h.Write(jsonData)
	signature := h.Sum(nil)

	// Combine signed data and original data (separated by ":")
	combinedData := append(signature, ':')
	combinedData = append(combinedData, jsonData...)

	// Encode combined data for storage in the cookie
	return base64.StdEncoding.EncodeToString(combinedData), nil
}

func VerifyData(signedData string) (map[string]string, error) {
	// Decode combined data
	decodedData, err := base64.StdEncoding.DecodeString(signedData)
	if err != nil {
		return nil, err
	}

	// Split data based on separator (":")
	parts := strings.SplitN(string(decodedData), ":", 2)
	if len(parts) != 2 {
		return nil, errors.New("invalid signed data format")
	}

	signature := []byte(parts[0])
	jsonData := []byte(parts[1])

	// Verify signature using secret key and session salt
	h := hmac.New(sha256.New, []byte(globals.Secret+globals.SessionSalt))
	h.Write(jsonData)
	expectedSignature := h.Sum(nil)

	if !hmac.Equal(signature, expectedSignature) {
		return nil, errors.New("invalid signature")
	}

	// Unmarshal JSON data
	var userData map[string]string
	if err := json.Unmarshal(jsonData, &userData); err != nil {
		return nil, err
	}

	return userData, nil
}
