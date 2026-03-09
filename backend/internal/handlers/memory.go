package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"openade/internal/model"
)

func (s *Server) HandleGetMemory(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "task_id")

	entries, err := s.Memory.GetAll(r.Context(), taskID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "get_failed", err.Error())
		return
	}
	if entries == nil {
		entries = []model.MemoryEntry{}
	}

	// Also return as a map for convenience
	entryMap := make(map[string]string)
	for _, e := range entries {
		entryMap[e.Key] = e.Value
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"entries": entryMap,
	})
}

func (s *Server) HandleSetMemory(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "task_id")

	var req model.SetMemoryRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "invalid request body")
		return
	}

	if err := s.Memory.SetAll(r.Context(), taskID, req.Entries); err != nil {
		writeError(w, http.StatusInternalServerError, "set_failed", err.Error())
		return
	}
	writeOK(w)
}

func (s *Server) HandleSetMemoryKey(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "task_id")
	key := chi.URLParam(r, "key")

	var req model.SetMemoryRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "invalid request body")
		return
	}

	if err := s.Memory.Set(r.Context(), taskID, key, req.Value); err != nil {
		writeError(w, http.StatusInternalServerError, "set_failed", err.Error())
		return
	}
	writeOK(w)
}
