package handlers

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"openade/internal/model"
)

func (s *Server) HandleListMCPServers(w http.ResponseWriter, r *http.Request) {
	servers, err := s.MCPServers.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list_failed", err.Error())
		return
	}
	if servers == nil {
		servers = []model.MCPServer{}
	}
	writeJSON(w, http.StatusOK, servers)
}

func (s *Server) HandleCreateMCPServer(w http.ResponseWriter, r *http.Request) {
	var req model.CreateMCPServerRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "invalid request body")
		return
	}

	server, err := s.MCPServers.Create(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "create_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, server)
}

func (s *Server) HandleUpdateMCPServer(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req model.UpdateMCPServerRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "invalid request body")
		return
	}

	server, err := s.MCPServers.Update(r.Context(), id, req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "not_found", err.Error())
			return
		}
		writeError(w, http.StatusBadRequest, "update_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, server)
}

func (s *Server) HandleDeleteMCPServer(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := s.MCPServers.Delete(r.Context(), id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "not_found", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "delete_failed", err.Error())
		return
	}
	writeOK(w)
}

func (s *Server) HandleTestMCPServer(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	result, err := s.MCPServers.Test(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "not_found", err.Error())
			return
		}
		writeError(w, http.StatusBadRequest, "test_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}
