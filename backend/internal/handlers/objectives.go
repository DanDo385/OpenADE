package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"openade/internal/model"
)

func (s *Server) HandleGetObjective(w http.ResponseWriter, r *http.Request) {
	convID := chi.URLParam(r, "id")

	obj, err := s.Objectives.GetByConversationID(r.Context(), convID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "get_failed", err.Error())
		return
	}
	if obj == nil {
		writeError(w, http.StatusNotFound, "not_found", "no objective for this conversation")
		return
	}

	writeJSON(w, http.StatusOK, obj)
}

func (s *Server) HandleUpsertObjective(w http.ResponseWriter, r *http.Request) {
	convID := chi.URLParam(r, "id")

	var req model.UpsertObjectiveRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "invalid request body")
		return
	}

	obj, err := s.Objectives.Upsert(r.Context(), convID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "upsert_failed", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, obj)
}

func (s *Server) HandleExportObjectiveMarkdown(w http.ResponseWriter, r *http.Request) {
	convID := chi.URLParam(r, "id")

	obj, err := s.Objectives.GetByConversationID(r.Context(), convID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "get_failed", err.Error())
		return
	}
	if obj == nil {
		writeError(w, http.StatusNotFound, "not_found", "no objective for this conversation")
		return
	}

	md := s.Objectives.ExportMarkdown(obj)

	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(md))
}
