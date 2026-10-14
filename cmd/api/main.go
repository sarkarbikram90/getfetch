package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bikrams/getfetch/internal/api"
	"github.com/bikrams/getfetch/internal/store/postgres"
)

func main() {
	var (
		port  = flag.Int("port", 8080, "HTTP server port")
		dbURL = flag.String("db", "", "PostgreSQL database URL (defaults to DATABASE_URL env)")
	)
	flag.Parse()

	databaseURL := *dbURL
	if databaseURL == "" {
		databaseURL = os.Getenv("DATABASE_URL")
	}
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgrespassword@localhost:5432/getfetch?sslmode=disable"
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	fmt.Println("===================================================================")
	fmt.Println("  getfetch / KubeJobs — REST API Server")
	fmt.Println("===================================================================")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	logger.Info("Connecting to PostgreSQL...", "url", databaseURL)
	store, err := postgres.NewStore(ctx, databaseURL)
	if err != nil {
		logger.Error("Failed to connect to database", "err", err)
		os.Exit(1)
	}
	defer store.Close()

	if err := store.Migrate(ctx); err != nil {
		logger.Error("Failed to apply migrations", "err", err)
		os.Exit(1)
	}
	logger.Info("Database migrations verified")

	server := api.NewServer(store, logger)

	addr := fmt.Sprintf(":%d", *port)
	httpServer := &http.Server{
		Addr:         addr,
		Handler:      server,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("REST API server listening", "addr", addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server failed", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	logger.Info("Shutting down server gracefully...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server shutdown failed", "err", err)
	}
	logger.Info("Server exited cleanly")
}
