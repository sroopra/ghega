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

	"github.com/sroopra/ghega/internal/server"
	"github.com/sroopra/ghega/pkg/messagestore"
)

func runServe(args []string) error {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	port := fs.Int("port", 8080, "HTTP server port")
	_ = fs.Parse(args)

	if envPort := os.Getenv("GHEGA_PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			*port = p
		}
	}

	store, err := initStore()
	if err != nil {
		return fmt.Errorf("init store: %w", err)
	}

	srv := server.New(store)
	httpSrv := &http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: srv.Handler(),
	}

	errCh := make(chan error, 1)
	go func() {
		fmt.Printf("Ghega server listening on port %d\n", *port)
		errCh <- httpSrv.ListenAndServe()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("server error: %w", err)
		}
	case sig := <-sigCh:
		fmt.Printf("\nReceived signal %s, shutting down...\n", sig)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
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
