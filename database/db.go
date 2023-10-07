//* database/db.go

package database

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"

	"aarushishop/globals"
)

var db *pgx.Conn

// ConnectToDB establishes a connection to the PostgreSQL database.
func ConnectToDB() (*pgx.Conn, error) {
	if db != nil {
		// If a connection already exists, return it
		return db, nil
	}

	config, err := globals.LoadConfig()
	if err != nil {
		return nil, errors.Wrap(err, "failed to load configuration")
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.DBHost, config.DBPort, config.DBUser, config.DBPassword, config.DBName, config.DBSSLMode)
	log.Printf("connection : %s", connStr)

	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to the database")
	}

	db = conn
	log.Printf("Connected to the database")
	return db, nil
}

// GetDB returns the PostgreSQL database connection.
func GetDB() *pgx.Conn {
	return db
}

// CloseDB closes the PostgreSQL database connection.
func CloseDB() error {
	if db != nil {
		if err := db.Close(context.Background()); err != nil {
			log.Printf("Failed to close the database connection: %v", err)
			return err
		}
		log.Printf("Closed the database connection")
	}
	return nil
}
