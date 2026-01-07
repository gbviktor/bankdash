package httpx

import (
	"encoding/json"
	"net/http"

	"bankdash/backend/internal/domain"
)

func (s *Server) handleListTemplates(w http.ResponseWriter, r *http.Request) {
	list, err := s.meta.ListTemplates()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, list, 200)
}

func (s *Server) handleUpsertCSVTemplate(w http.ResponseWriter, r *http.Request) {
	var t domain.BankTemplate
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	if t.Type == "" {
		t.Type = "csv"
	}
	if err := s.meta.UpsertTemplate(t); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	writeJSON(w, map[string]any{"ok": true, "id": t.ID}, 200)
}

func writeJSON(w http.ResponseWriter, v any, status int) {
	w.Header().Set("content-type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
