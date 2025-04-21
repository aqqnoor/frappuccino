package database

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	_ "github.com/lib/pq"
)

// Connect opens a connection to the database with retries
func Connect(dsn string, logger *slog.Logger) (*sql.DB, error) {
	const op = "database.Connect"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		logger.Error(fmt.Sprintf("%s: Error creating sql.DB", op), "Error", err)
		return nil, err
	}

	logger.Info("PING: Trying to connect to the database")
	for i := 1; i < 6; i++ {
		if err := db.Ping(); err == nil {
			logger.Info("PING: Successfully connected to the database")
			return db, nil
		}

		logger.Warn(fmt.Sprintf("PING: Error connecting to the database: %v", err))
		if i == 5 {
			break
		}
		logger.Info(fmt.Sprintf("PING: Trying to connect again #%d...", i))
		time.Sleep(3 * time.Second)
	}

	logger.Error("PING: Failed to connect to database after 5 attempts")
	return nil, fmt.Errorf("connection to the database is not established")
}
