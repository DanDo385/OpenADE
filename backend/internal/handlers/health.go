package handlers

import "net/http"

// HandleHealth returns a simple health check response.
// Tauri polls this endpoint on startup to confirm the backend is ready.
func (s *Server) HandleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
