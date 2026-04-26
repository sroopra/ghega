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

	"github.com/sroopra/ghega/internal/engine"
	"github.com/sroopra/ghega/pkg/messagestore"
	"github.com/sroopra/ghega/pkg/mllp"
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

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	// Initialize message store: prefer SQLite, fall back to MemoryStore.
	var store messagestore.Store
	dsn := os.Getenv("GHEGA_DATABASE_URL")
	if dsn == "" {
		dsn = "ghega.db"
	}
	sqliteStore, err := messagestore.NewSQLiteStore(dsn)
	if err != nil {
		logger.Warn("failed to open sqlite store, falling back to memory store", slog.String("error", err.Error()))
		store = messagestore.NewInMemoryStore()
	} else {
		store = sqliteStore
		logger.Info("sqlite store initialized", slog.String("dsn", dsn))
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

	// Initialize MLLP listener.
	eng := engine.NewEngine(store, logger)
	mllpCfg := mllp.ConfigFromEnv()
	mllpListener := mllp.NewListener(mllpCfg, eng.MLLPHandler(), logger)

	if err := mllpListener.Start(); err != nil {
		return fmt.Errorf("mllp listener start error: %w", err)
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("ghega http server listening", slog.Int("port", *port))
		errCh <- srv.ListenAndServe()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			_ = mllpListener.Stop()
			return fmt.Errorf("server error: %w", err)
		}
	case sig := <-sigCh:
		logger.Info("received shutdown signal", slog.String("signal", sig.String()))
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
		_ = mllpListener.Stop()
		return nil
	}

	return nil
}
