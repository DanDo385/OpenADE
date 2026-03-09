package handlers

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"openade/internal/model"
)

func (s *Server) HandleListAgents(w http.ResponseWriter, r *http.Request) {
	agents, err := s.Agents.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list_failed", err.Error())
		return
	}
	if agents == nil {
		agents = []model.Agent{}
	}
	writeJSON(w, http.StatusOK, agents)
}

func (s *Server) HandleGetAgent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	agent, err := s.Agents.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "get_failed", err.Error())
		return
	}
	if agent == nil {
		writeError(w, http.StatusNotFound, "not_found", "agent not found")
		return
	}
	writeJSON(w, http.StatusOK, agent)
}

func (s *Server) HandleGetAgentBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	agent, err := s.Agents.GetBySlug(r.Context(), slug)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "get_failed", err.Error())
		return
	}
	if agent == nil {
		writeError(w, http.StatusNotFound, "not_found", "agent not found")
		return
	}
	writeJSON(w, http.StatusOK, agent)
}

func (s *Server) HandleRunAgent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req model.AgentRunRequest
	if err := decodeJSON(r, &req); err != nil {
		req = model.AgentRunRequest{}
	}

	resp, err := s.Agents.Run(r.Context(), id, req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "not_found", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "run_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, resp)
}
