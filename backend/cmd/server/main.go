// Command server runs the leoforio API.
//
// All upstream access goes through the OASA Telematics adapter, behind caches
// that keep the upstream load proportional to distinct stops and routes
// rather than to the number of clients. The static stop index comes from the
// official GTFS feed, imported on a schedule.
//
// See docs/data-sources/README.md for the source policy and
// docs/adr/0004-app-never-calls-upstream.md for why the app never talks to
// an upstream directly.
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
	"github.com/ksohos333/leoforio/backend/internal/gtfs"
	"github.com/ksohos333/leoforio/backend/internal/source/oasa"
	"github.com/ksohos333/leoforio/backend/internal/store"
)

// importInterval is how often the static feed is checked for changes.
const importInterval = 24 * time.Hour

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	addr := envOr("LEOFORIO_ADDR", ":8080")
	version := envOr("LEOFORIO_VERSION", "dev")
	databasePath := envOr("LEOFORIO_DB", "leoforio.db")

	localStore, err := store.Open(databasePath)
	if err != nil {
		logger.Error("cannot open the local store", "path", databasePath, "error", err)
		os.Exit(1)
	}
	defer localStore.Close()

	upstream := oasa.New(oasa.NewClient(envOr("LEOFORIO_OASA_BASE_URL", "")))
	logger.Info("upstream configured", "base_url", oasa.DefaultBaseURL)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go importLoop(ctx, logger, localStore)

	srv := &http.Server{
		Addr: addr,
		Handler: api.New(api.Dependencies{
			Arrivals:  upstream,
			Positions: upstream,
			Network:   upstream,
			Stops:     localStore,
			Version:   version,
		}).Routes(),
		ReadHeaderTimeout: 10 * time.Second,
	}

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

// importLoop keeps the local stop index up to date.
//
// A failed import is logged and retried at the next interval: the live
// features keep working without the static index, and the search endpoints
// report that the index is not available.
func importLoop(ctx context.Context, logger *slog.Logger, target *store.Store) {
	importer := gtfs.NewImporter(target, envOr("LEOFORIO_GTFS_URL", ""))

	importOnce := func() {
		started := time.Now()
		result, err := importer.Import(ctx)
		if err != nil {
			logger.Error("static feed import failed", "error", err)
			return
		}
		if result.Unchanged {
			logger.Info("static feed unchanged", "version", result.Version)
			return
		}
		logger.Info("static feed imported",
			"stops", result.Stops,
			"version", result.Version,
			"duration", time.Since(started).Round(time.Second).String())
	}

	importOnce()

	ticker := time.NewTicker(importInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			importOnce()
		}
	}
}

func envOr(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}
