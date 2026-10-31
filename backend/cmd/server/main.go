// Command server runs the leoforio API.
//
// Upstream sources are not wired yet: the endpoints that need one report
// 503 until a source is added. See docs/data-sources/README.md for the
// source policy and docs/adr/0004-app-never-calls-upstream.md for why all
// upstream access lives behind the API.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ksohos333/leoforio/backend/internal/api"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	addr := envOr("LEOFORIO_ADDR", ":8080")
	version := envOr("LEOFORIO_VERSION", "dev")

	srv := &http.Server{
		Addr:              addr,
		Handler:           api.New(nil, nil, version).Routes(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		logger.Info("leoforio api listening", "addr", addr, "version", version)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		logger.Error("server failed", "error", err)
		os.Exit(1)
	case <-ctx.Done():
		logger.Info("shutting down")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}
}

func envOr(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}
