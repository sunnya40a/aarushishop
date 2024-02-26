// globals/config.go

package globals

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"time"
)

// for strong session
var Charsets = ("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()")

var CopyrightYear = time.Now().Year()

// For session
var Secret = []byte("sCc6ef3WBLjZ3@rtqGhMhGCMGuDqYgfHS9Y&Pi5mSjpyfIgbsc_fg05Duc3x2dH4E9IfpyKlHXCy5XhiGF0s5A")

// var Charsets = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()")
var JwtSecretKey = []byte("ui5uJjqWVZxDd21vDLoB7_18tKRmFn&XYcyxuPw@XPAHt65bmfHzit8blm3c9G53QdOiAK1wRTxhjGY70lLQbw")
var RefreshSecretKey = []byte("51tEObTgkqhGvJhpT_MwspZwAMk&NiHb9DoWAKFFSq5d7G8rwy4tuKrOvRS2bP4S5+RrK@G1fJtR5ZpbYRBaMA")

const UserKey = "user"

// AppConfig stores the application configuration.
type AppConfig struct {
	DBHost     string `json:"DBHost"`
	DBPort     string `json:"DBPort"`
	DBUser     string `json:"DBUser"`
	DBPassword string `json:"DBPassword"`
	DBName     string `json:"DBName"`
	DBSSLMode  string `json:"DBSSLMode"`
}

// LoadConfig loads the application configuration from the config.json file.
func LoadConfig() (*AppConfig, error) {
	// Get the path to the config.json file (assuming it's in the same directory as the executable)
	configFile := filepath.Join("globals", "config.json")

	// Read the contents of the config file
	configData, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	// Parse the JSON data into the AppConfig struct
	var config AppConfig
	if err := json.Unmarshal(configData, &config); err != nil {
		return nil, err
	}

	log.Printf("==> %v", config)
	return &config, nil
}

func Copyright() int {
	return CopyrightYear
}
