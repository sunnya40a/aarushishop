// database/db.go

package database

import (
	"context"
	"database/sql"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql" // Import the MySQL driver
	"github.com/pkg/errors"
)

var db *sql.DB

// InitDB initializes the database connection.
func InitDB() error {
	connStr := "chhabi:NewPassword@tcp(localhost:3306)/shop?parseTime=true" // Use TCP for MariaDB

	// Create a new database connection
	var err error
	db, err = sql.Open("mysql", connStr)
	if err != nil {
		return errors.Wrap(err, "failed to open database connection")
	}

	// Ping the database to check the connection
	err = db.PingContext(context.Background())
	if err != nil {
		return errors.Wrap(err, "failed to ping database")
	}

	// Set connection pool size and idle timeout (optional)
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(5 * time.Minute)

	log.Printf("Connected to MariaDB database")
	return nil
}

// GetDB returns the database connection.
func GetDB() *sql.DB {
	return db
}

// CloseDB closes the database connection.
func CloseDB() {
	if db != nil {
		db.Close()
		log.Printf("Closed the database connection")
	}
}
