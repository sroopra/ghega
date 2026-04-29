package cli

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/sroopra/ghega/internal/alerts"
	"github.com/sroopra/ghega/internal/config"
	"github.com/sroopra/ghega/internal/engine"
	"github.com/sroopra/ghega/internal/server"
	"github.com/sroopra/ghega/pkg/channelstore"
	"github.com/sroopra/ghega/pkg/messagestore"
	"github.com/sroopra/ghega/pkg/mllp"
)

func runServe(args []string) error {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	port := fs.Int("port", 8080, "HTTP server port")
	migrationsDir := fs.String("migrations-dir", "", "Directory containing migration reports")
	_ = fs.Parse(args)

	if envPort := os.Getenv("GHEGA_PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			*port = p
		}
	}

	if *migrationsDir == "" {
		*migrationsDir = config.MigrationsDir("")
	}

	store, err := initStore()
	if err != nil {
		return fmt.Errorf("init store: %w", err)
	}

	chStore, err := initChannelStore()
	if err != nil {
		return fmt.Errorf("init channel store: %w", err)
	}

	// Start HTTP API server.
	alertStore := alerts.NewInMemoryAlertStore()
	srv := server.New(store, alertStore, chStore)
	srv.SetMigrationsDir(*migrationsDir)
	httpSrv := &http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: srv.Handler(),
	}

	httpErrCh := make(chan error, 1)
	go func() {
		fmt.Printf("Ghega HTTP server listening on port %d\n", *port)
		httpErrCh <- httpSrv.ListenAndServe()
	}()

	// Start MLLP listener.
	mllpCfg := mllp.ConfigFromEnv()
	mllpHandler := engine.NewMLLPHandler(store, engine.DefaultHandlerConfig(), slog.Default())
	mllpListener := mllp.NewListener(mllpCfg, mllpHandler, slog.Default())

	if err := mllpListener.Start(); err != nil {
		slog.Warn("failed to start MLLP listener; HTTP server will still run", slog.String("error", err.Error()))
	} else {
		fmt.Printf("Ghega MLLP listener listening on %s:%d\n", mllpCfg.Host, mllpCfg.Port)
	}

	// Wait for shutdown signal.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-httpErrCh:
		if err != nil && err != http.ErrServerClosed {
			if mllpListener.Addr() != nil {
				_ = mllpListener.Stop()
			}
			return fmt.Errorf("http server error: %w", err)
		}
	case sig := <-sigCh:
		fmt.Printf("\nReceived signal %s, shutting down...\n", sig)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if mllpListener.Addr() != nil {
			_ = mllpListener.Stop()
		}
		return httpSrv.Shutdown(ctx)
	}

	return nil
}

func initStore() (messagestore.Store, error) {
	dsn := os.Getenv("GHEGA_DATABASE_URL")
	if dsn == "" {
		dsn = "ghega.db"
	}

	store, err := messagestore.NewSQLiteStore(dsn)
	if err != nil {
		slog.Warn("failed to open SQLite store, falling back to in-memory store", "error", err)
		return messagestore.NewInMemoryStore(), nil
	}
	return store, nil
}

func initChannelStore() (channelstore.ChannelStore, error) {
	dsn := os.Getenv("GHEGA_DATABASE_URL")
	if dsn == "" {
		dsn = "ghega.db"
	}

	store, err := channelstore.NewSQLiteStore(dsn)
	if err != nil {
		slog.Warn("failed to open SQLite channel store, falling back to in-memory store", "error", err)
		return channelstore.NewInMemoryStore(), nil
	}
	return store, nil
}
