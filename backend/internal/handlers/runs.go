package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"openade/internal/model"
)

func (s *Server) HandleListRuns(w http.ResponseWriter, r *http.Request) {
	taskID := r.URL.Query().Get("task_id")
	runs, err := s.Runs.List(r.Context(), taskID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list_failed", err.Error())
		return
	}
	if runs == nil {
		runs = []model.Run{}
	}
	writeJSON(w, http.StatusOK, runs)
}

func (s *Server) HandleGetRun(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	run, err := s.Runs.Get(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "get_failed", err.Error())
		return
	}
	if run == nil {
		writeError(w, http.StatusNotFound, "not_found", "run not found")
		return
	}
	writeJSON(w, http.StatusOK, run)
}
