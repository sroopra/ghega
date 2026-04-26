package cli

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/sroopra/ghega/internal/api"
	"github.com/sroopra/ghega/pkg/messagestore"
)

func runServe(args []string) error {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	port := fs.Int("port", 8080, "HTTP server port")
	devAuth := fs.Bool("dev-auth", false, "Bypass authentication for local development")
	_ = fs.Parse(args)

	if envPort := os.Getenv("GHEGA_PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			*port = p
		}
	}

	// Use an in-memory message store for now. In production this would be
	// configurable (SQLite, etc.).
	msgStore := messagestore.NewInMemoryStore()
	chStore := api.NewInMemoryChannelStore()

	handler := api.NewHandler(msgStore, chStore)
	router := api.NewRouter(handler, *devAuth)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: router,
	}

	errCh := make(chan error, 1)
	go func() {
		fmt.Printf("Ghega server listening on port %d\n", *port)
		errCh <- srv.ListenAndServe()
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
		return srv.Shutdown(ctx)
	}

	return nil
}
