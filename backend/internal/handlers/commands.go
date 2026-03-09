package handlers

import (
	"net/http"

	"openade/internal/model"
)

func (s *Server) HandleExecuteCommand(w http.ResponseWriter, r *http.Request) {
	var req model.CommandExecuteRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "invalid request body")
		return
	}

	resp := s.Commands.Execute(r.Context(), req)
	writeJSON(w, http.StatusOK, resp)
}
