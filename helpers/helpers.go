package helpers

import (
	"log"
	"strings"
)

// CheckUserPass checks if the given username and password are valid
func CheckUserPass(username, password string) bool {
	// Define a map to store username-password pairs
	userpass := make(map[string]string)
	userpass["chhabi"] = "1234"
	userpass["admin"] = "admin123"

	log.Println("checkUserPass", username, password, userpass)

	// Check if the username exists in the map
	if val, ok := userpass[username]; ok {
		log.Println(val, ok)
		// If the username exists, compare the password
		if val == password {
			return true
		} else {
			return false
		}
	} else {
		return false
	}
}

// EmptyUserPass checks if the given username and password are empty
func EmptyUserPass(username, password string) bool {
	return strings.Trim(username, " ") == "" || strings.Trim(password, " ") == ""
}
