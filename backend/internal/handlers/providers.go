package handlers

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"openade/internal/model"
)

func (s *Server) HandleListProviders(w http.ResponseWriter, r *http.Request) {
	configs, err := s.Providers.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list_failed", err.Error())
		return
	}
	if configs == nil {
		configs = []model.ProviderConfig{}
	}

	// Strip API keys from response — only show configured status
	safe := make([]map[string]interface{}, len(configs))
	for i, c := range configs {
		safe[i] = map[string]interface{}{
			"id":            c.ID,
			"provider":      c.Provider,
			"configured":    c.Configured,
			"default_model": c.DefaultModel,
			"base_url":      c.BaseURL,
		}
	}
	writeJSON(w, http.StatusOK, safe)
}

func (s *Server) HandleSaveProvider(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "id")

	var req model.SaveProviderRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "invalid request body")
		return
	}
	if strings.TrimSpace(req.APIKey) == "" {
		writeError(w, http.StatusBadRequest, "missing_key", "api_key is required")
		return
	}

	cfg, err := s.Providers.Save(r.Context(), provider, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "save_failed", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"provider":      cfg.Provider,
		"configured":    cfg.Configured,
		"default_model": cfg.DefaultModel,
	})
}
