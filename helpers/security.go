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
// For strings, it first sanitizes against SQL injection using sanitizeSQL,
// For time.Time objects, it returns them as is.
// Unsupported types are returned as nil.
func SanitizeData(input interface{}) interface{} {
	// bluemonday.StrictPolicy() which can be thought of as equivalent to stripping all HTML elements and their attributes as it has nothing on its allowlist. An example usage scenario would be blog post titles where HTML tags are not expected at all and if they are then the elements and the content of the elements should be stripped. This is a very strict policy.
	// bluemonday.UGCPolicy() which allows a broad selection of HTML elements and attributes that are safe for user generated content. Note that this policy does not allow iframes, object, embed, styles, script, etc. An example usage scenario would be blog post bodies where a variety of formatting is expected along with the potential for TABLEs and IMGs.

	p := bluemonday.StrictPolicy()
	//AntiSamy inherently includes Bluemonday's features as its core engine. future plan

	switch v := input.(type) {
	case int:
		// For integers, just return as is
		return v
	case string:
		// For strings, sanitize against SQL injection and XSS attacks
		// SQL Injection protection
		//cleanString := sanitizeSQL(v)
		cleanString := p.Sanitize(v)
		cleanString = SanitizeSQL(cleanString, "mysql")
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

// sanitizeSQL sanitizes input string against SQL injection.
// It replaces single quotes with double single quotes to escape them.
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

// below function not required as we have used sanitize.html function after caling this function.
func sanitizeInput(input string) string {
	// Replace HTML entities with their respective characters
	input = html.UnescapeString(input)

	// Define regular expression pattern to allow only alphanumeric characters, space, dash, digits, and parentheses
	regexPattern := regexp.MustCompile(`[^\s\w\-\(\)!]`)
	//[^\s\w\-\(\)!] matches any character that is not a whitespace (\s), word character (\w), hyphen (\-), left parenthesis (\(), right parenthesis (\)), or exclamation mark (!).

	// Replace any characters not matching the pattern with an empty string
	sanitizedInput := regexPattern.ReplaceAllString(input, "")

	return sanitizedInput
}
