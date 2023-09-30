// globals/config.go

package globals

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

// For session
var Secret = []byte("YUcD6G8qzz/zwb5nxd6Z1/Uj8x7Q5F1C+JALBfEfjZEYfhYSLyrCVBS/uxWxmESA")

const Userkey = "user"

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
	// Get the path to the config.json file (assuming it's in the globals folder)
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
