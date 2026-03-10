package handlers

import "net/http"

// HandleRoot returns a simple API info response for GET /.
// Avoids 404 when users or tooling hit the backend root.
func (s *Server) HandleRoot(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"service": "OpenADE API",
		"health":  "/health",
		"api":     "/api",
	})
}

// HandleHealth returns a simple health check response.
// Tauri polls this endpoint on startup to confirm the backend is ready.
func (s *Server) HandleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
