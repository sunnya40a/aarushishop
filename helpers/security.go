// Package helpers provides utility functions for sanitizing data against SQL injection and XSS attacks.
package helpers

import (
	"html"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/microcosm-cc/bluemonday"
)

// SanitizeData sanitizes input data against SQL injection and XSS attacks.
// For integers, it returns them as is.
// For strings, it first sanitizes against SQL injection using SanitizeSQL,
// then sanitizes using bluemonday.StrictPolicy() to prevent XSS attacks.
// For time.Time objects, it returns them as is.
// Unsupported types are returned as nil.
func SanitizeData(input interface{}) interface{} {
	p := bluemonday.StrictPolicy()

	switch v := input.(type) {
	case int:
		// For integers, just return as is
		return v
	case string:
		// For strings, sanitize against SQL injection and XSS attacks
		cleanString := SanitizeSQL(v, "mysql")
		cleanString = p.Sanitize(cleanString)
		cleanString = sanitizeInput(cleanString)
		return cleanString
	case time.Time:
		// For dates, just return as is
		return v
	default:
		// For unsupported types, return nil
		return nil
	}
}

// SanitizeSQL sanitizes input string against SQL injection.
// It replaces single quotes with double single quotes to escape them.
// It takes the input string and the database dialect as parameters.
func SanitizeSQL(input string, dialect string) string {
	var sanitized strings.Builder
	var lastChar rune

	// Escape characters based on provided dialect
	escapeMap := map[string]map[rune]rune{
		"mysql": {
			'\'': '\'', // Escape single quotes for MySQL
			';':  '\\', // Escape semicolons for MySQL
			'\\': '\\', // Escape backslashes for MySQL
		},

		"postgres": {
			'\'': '\'', // Escape single quotes for PostgreSQL
			';':  '\\', // Escape semicolons for PostgreSQL
			'\\': '\\', // Escape backslashes for PostgreSQL
		},
		// Add mapping for other dialects as needed
	}

	for _, char := range input {
		escapedChar, ok := escapeMap[dialect][char]
		if ok {
			sanitized.WriteRune(escapedChar)
			lastChar = escapedChar
		} else if unicode.IsLetter(char) || unicode.IsNumber(char) || char == '_' || char == ' ' || char == '.' || char == ',' || char == '-' || char == '/' {
			sanitized.WriteRune(char)
			lastChar = char
		} else {
			if lastChar != '\\' {
				sanitized.WriteRune('\\')
			}
			sanitized.WriteRune(char)
			lastChar = char
		}
	}

	return sanitized.String()
}

// sanitizeInput sanitizes input string by removing non-alphanumeric characters.
// It replaces HTML entities with their respective characters.
// It returns the sanitized input string.
func sanitizeInput(input string) string {
	// Replace HTML entities with their respective characters
	input = html.UnescapeString(input)
	// Define regular expression pattern to allow only alphanumeric characters, space, dash, digits, and parentheses
	regexPattern := regexp.MustCompile(`[^a-zA-Z0-9()\-!\[\]/\s]`)
	// [^a-zA-Z0-9()\-!\[\]/\s] matches any character that is not alphanumeric, space, dash, parentheses, exclamation mark, or square brackets.

	// Replace any characters not matching the pattern with an empty string
	sanitizedInput := regexPattern.ReplaceAllString(input, "")
	// Replace multiple spaces with a empty string
	sanitizedInput = strings.Join(strings.Fields(sanitizedInput), " ")
	// Replace multiple spaces with a single space
	return sanitizedInput
}

// SanitizeUsername sanitizes input username against SQL injection and XSS attacks.
// It first sanitizes against SQL injection using SanitizeSQL,
// then sanitizes using bluemonday.StrictPolicy() to prevent XSS attacks.
// It returns the sanitized username.
func SanitizeUsername(input string) string {
	p := bluemonday.StrictPolicy()
	// Sanitize against SQL injection and XSS attacks
	cleanString := SanitizeSQL(input, "mysql")
	cleanString = p.Sanitize(cleanString)
	// Clean username by removing non-alphanumeric characters
	cleanString = cleanUsername(cleanString)
	return cleanString
}

// SanitizePassword sanitizes input password against SQL injection and XSS attacks.
// then sanitizes using bluemonday.StrictPolicy() to prevent XSS attacks.
// It returns the sanitized password.
func SanitizePassword(input string) string {
	cleanString := cleanPassword(input)
	return cleanString
}

// cleanUsername sanitizes input username by removing non-alphanumeric characters.
// It replaces HTML entities with their respective characters.
// It returns the sanitized username.
func cleanUsername(input string) string {
	// Replace HTML entities with their respective characters
	input = html.UnescapeString(input)
	// Define regular expression pattern to allow only alphanumeric characters
	regexPattern := regexp.MustCompile(`[^a-zA-Z0-9.]`)
	// Replace any characters not matching the pattern with an empty string
	sanitizedInput := regexPattern.ReplaceAllString(input, "")
	sanitizedInput = strings.Join(strings.Fields(sanitizedInput), " ")
	return sanitizedInput
}

// sanitizePassword removes unnecessary characters from a password
func cleanPassword(password string) string {
	// Define a regular expression pattern to match allowed characters
	allowedCharsPattern := regexp.MustCompile(`[^a-zA-Z0-9!"#$%&'()*+,\-./:;<=>?@[\\]^_{|}~]`)
	// Remove disallowed characters from the password
	cleanPassword := allowedCharsPattern.ReplaceAllString(password, "")
	return cleanPassword
}
