package server

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Server struct {
	httpServer *http.Server
	logger     *slog.Logger
}

func NewServer(db *sql.DB, logger *slog.Logger) *Server {
	router := http.NewServeMux()

	setupRoutes(router, db, logger)

	return &Server{
		httpServer: &http.Server{
			Addr:         ":8080",
			Handler:      router,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
		logger: logger,
	}
}

func (s *Server) Start() {
	const op = "server.Server.Start"
	s.logger.Info("Server launched at http://localhost:8080")
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil {
			s.logger.Error(fmt.Sprintf("%s: Error starting server", op), "error", err)
			os.Exit(1)
		}
	}()
}

func (s *Server) Shutdown() {
	const op = "server.Server.Shutdown"
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	s.logger.Info(fmt.Sprintf("%s: Termination signal received. Stopping server...", op))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.Error(fmt.Sprintf("%s: Error while stopping the server", op), "error", err)
	}
	s.logger.Info(fmt.Sprintf("%s: Server stopped successfully", op))
}
