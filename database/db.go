// database/db.go

// * postgres://chhabi:0hzwAcutBoSeC0rqV0bQ8PVbsB5at1TL@dpg-ck673fj6fquc73ddjncg-a.singapore-postgres.render.com:5432/school_y6lf?sslmode=prefer

package database

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

var dbPool *pgxpool.Pool

// InitDBPool initializes the database connection pool.
func InitDBPool() error {
	connStr := "postgres://chhabi:0hzwAcutBoSeC0rqV0bQ8PVbsB5at1TL@dpg-ck673fj6fquc73ddjncg-a.singapore-postgres.render.com:5432/school_y6lf?sslmode=prefer"

	// Create a new connection pool configuration
	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return errors.Wrap(err, "failed to parse connection string")
	}

	// Customize connection pool size (adjust as needed)
	config.MaxConns = 5 // Set your desired maximum connection pool size

	// Customize connection pool cleanup (set idle timeout)
	config.MaxConnIdleTime = 5 * time.Minute // Close idle connections after 5 minutes

	// Create a new connection pool using the parsed configuration
	dbPool, err = pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return errors.Wrap(err, "failed to connect to the database")
	}

	log.Printf("Connected to the database")
	return nil
}

// GetDBConnection returns a connection from the database pool.
func GetDBConnection() (*pgxpool.Conn, error) {
	conn, err := dbPool.Acquire(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "failed to acquire database connection")
	}
	return conn, nil
}

// CloseDBPool closes the database connection pool.
func CloseDBPool() {
	if dbPool != nil {
		dbPool.Close()
		log.Printf("Closed the database connection pool")
	}
}
