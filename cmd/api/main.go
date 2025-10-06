package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Nassabiq/gpci-compro-api/internal/app"
	"github.com/Nassabiq/gpci-compro-api/internal/config"
	"github.com/pressly/goose/v3"
)

// Migrations are loaded from disk under ./migrations when present.

func main() {
	cfg := config.Load()

	ctx := context.Background()
	container, cleanup, err := app.New(ctx, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize container: %v\n", err)
		os.Exit(1)
	}
	defer cleanup()

	container.Logger.Info("starting api", "env", cfg.App.Env, "port", cfg.App.Port)

	if err := runMigrations(container.DB); err != nil {
		container.Logger.Error("migrate", "err", err)
		os.Exit(1)
	}

	fiberApp := app.Setup(container)

	go func() {
		if err := fiberApp.Listen(":" + cfg.App.Port); err != nil {
			container.Logger.Error("listen", "err", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.App.ShutdownTimeout)
	defer cancel()

	_ = fiberApp.ShutdownWithContext(shutdownCtx)
	container.Logger.Info("server stopped")
}

func runMigrations(database *sql.DB) error {
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	entries, err := os.ReadDir("migrations")
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if len(entries) == 0 {
		return nil
	}
	return goose.Up(database, "migrations")
}
