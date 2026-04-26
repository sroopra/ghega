package cli

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/sroopra/ghega/internal/channel"
	"github.com/sroopra/ghega/internal/logging"
	"github.com/sroopra/ghega/pkg/httpsender"
	"github.com/sroopra/ghega/pkg/messagestore"
)

func runServe(args []string) error {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	port := fs.Int("port", 8080, "HTTP server port")
	channelsPath := fs.String("channels", "", "Path to channel YAML config to start an integration channel")
	_ = fs.Parse(args)

	if envPort := os.Getenv("GHEGA_PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			*port = p
		}
	}

	var ch *channel.Channel
	if *channelsPath != "" {
		cfg, err := channel.LoadConfig(*channelsPath)
		if err != nil {
			return fmt.Errorf("load channel config: %w", err)
		}

		logger := logging.New(os.Stderr, slog.LevelInfo)
		store := messagestore.NewInMemoryStore()
		sender := &httpsender.Sender{
			URL:     cfg.Destination.URL,
			Timeout: 30 * time.Second,
			Retries: 3,
		}

		ch = channel.NewChannel(*cfg, store, sender, logger)
		if err := ch.Run(); err != nil {
			return fmt.Errorf("start channel: %w", err)
		}
		fmt.Printf("Ghega channel %q started on %s\n", cfg.Name, ch.Addr())
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: mux,
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
		if ch != nil {
			_ = ch.Stop()
		}
		return srv.Shutdown(ctx)
	}

	return nil
}
