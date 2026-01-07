package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"bankdash/backend/internal/config"
	"bankdash/backend/internal/httpx"
	"bankdash/backend/internal/influx"
	"bankdash/backend/internal/meta"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("config load failed")
	}

	metaStore, err := meta.Open(cfg.MetaDBPath)
	if err != nil {
		log.Fatal().Err(err).Msg("meta store open failed")
	}
	defer metaStore.Close()

	influxClient, err := influx.New(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("influx client init failed")
	}
	defer influxClient.Close()

	// seed templates from /app/config/templates (mounted by compose)
	if err := metaStore.SeedTemplatesFromDir(cfg.TemplateDir); err != nil {
		log.Warn().Err(err).Msg("template seeding failed (continuing)")
	}

	srv := httpx.NewServer(cfg, metaStore, influxClient)

	httpServer := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           srv.Router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Info().Str("addr", httpServer.Addr).Msg("server started")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server crashed")
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = httpServer.Shutdown(ctx)
	log.Info().Msg("server stopped")
}
