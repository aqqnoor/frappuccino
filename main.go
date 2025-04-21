package main

import (
	"log/slog"
	"os"

	"github.com/sunzhqr/frappuccino/config"
	"github.com/sunzhqr/frappuccino/internal/server"
	"github.com/sunzhqr/frappuccino/pkg/database"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	dsn := config.GetDBConfig()

	db, err := database.Connect(dsn, logger)
	if err != nil {
		logger.Error("Database connection error", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	srv := server.NewServer(db, logger)
	srv.Start()

	go srv.Shutdown()

	select {}
}
