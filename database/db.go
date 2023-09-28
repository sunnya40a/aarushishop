package database

import (
	"aarushishop/globals"
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
)

// ConnectToDB establishes a connection to the PostgreSQL database.
func ConnectToDB() (*pgx.Conn, error) {
	config, err := globals.LoadConfig()
	if err != nil {
		log.Printf("Error loading configuration: %v", err)
		return nil, err
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.DBHost, config.DBPort, config.DBUser, config.DBPassword, config.DBName, config.DBSSLMode)

	db, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		log.Printf("Error connecting to the database: %v", err)
		return nil, err
	}

	return db, nil
}
