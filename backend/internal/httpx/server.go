package httpx

import (
	"net/http"
	"time"

	"bankdash/backend/internal/config"
	"bankdash/backend/internal/influx"
	"bankdash/backend/internal/meta"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	Router *chi.Mux

	cfg   config.Config
	meta  *meta.Store
	inflx *influx.Client
}

func NewServer(cfg config.Config, metaStore *meta.Store, inflx *influx.Client) *Server {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	s := &Server{Router: r, cfg: cfg, meta: metaStore, inflx: inflx}

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	r.Route("/api/v1", func(api chi.Router) {
		api.Get("/templates", s.handleListTemplates)
		api.Post("/templates/csv", s.handleUpsertCSVTemplate)

		api.Post("/imports/csv", s.handleImportCSV)
	})

	return s
}
