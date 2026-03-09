package handlers

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"openade/internal/model"
)

func (s *Server) HandleCreateTask(w http.ResponseWriter, r *http.Request) {
	var req model.CreateTaskRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "invalid request body")
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		writeError(w, http.StatusBadRequest, "missing_name", "task name is required")
		return
	}
	if strings.TrimSpace(req.PromptTemplate) == "" {
		writeError(w, http.StatusBadRequest, "missing_template", "prompt template is required")
		return
	}

	task, err := s.Tasks.Create(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "create_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, task)
}

func (s *Server) HandleListTasks(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	tasks, err := s.Tasks.List(r.Context(), query)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list_failed", err.Error())
		return
	}
	if tasks == nil {
		tasks = []model.Task{}
	}
	writeJSON(w, http.StatusOK, tasks)
}

func (s *Server) HandleGetTask(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	task, err := s.Tasks.Get(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "get_failed", err.Error())
		return
	}
	if task == nil {
		writeError(w, http.StatusNotFound, "not_found", "task not found")
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func (s *Server) HandleUpdateTask(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req model.UpdateTaskRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "invalid request body")
		return
	}

	task, err := s.Tasks.Update(r.Context(), id, req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "not_found", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "update_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func (s *Server) HandleDeleteTask(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := s.Tasks.Delete(r.Context(), id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "not_found", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "delete_failed", err.Error())
		return
	}
	writeOK(w)
}

func (s *Server) HandleRunTask(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	ctx := r.Context()

	var req model.RunTaskRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "invalid request body")
		return
	}

	// Load task
	task, err := s.Tasks.Get(ctx, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "get_failed", err.Error())
		return
	}
	if task == nil {
		writeError(w, http.StatusNotFound, "not_found", "task not found")
		return
	}

	// Get LLM adapter
	adapter, provCfg, err := s.getLLMAdapter(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "provider_error", err.Error())
		return
	}
	if adapter == nil {
		writeError(w, http.StatusUnauthorized, "no_provider", "no LLM provider configured")
		return
	}

	// Resolve model
	runModel := req.Model
	if runModel == "" && provCfg != nil {
		runModel = provCfg.DefaultModel
	}

	// Execute run
	run, err := s.Runs.Execute(ctx, task, req.Inputs, adapter, runModel)
	if err != nil {
		writeError(w, http.StatusBadGateway, "run_failed", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, run)
}

func (s *Server) HandleExportTask(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	ctx := r.Context()

	bundle, err := s.Tasks.Export(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "not_found", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "export_failed", err.Error())
		return
	}

	// Attach memory
	memMap, _ := s.Memory.GetMap(ctx, id)
	bundle.Memory = memMap

	writeJSON(w, http.StatusOK, bundle)
}

func (s *Server) HandleImportTask(w http.ResponseWriter, r *http.Request) {
	var bundle model.ExportBundle
	if err := decodeJSON(r, &bundle); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "invalid import bundle")
		return
	}
	if strings.TrimSpace(bundle.Task.Name) == "" {
		writeError(w, http.StatusBadRequest, "missing_name", "task name is required in bundle")
		return
	}

	ctx := r.Context()
	task, err := s.Tasks.Import(ctx, bundle)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "import_failed", err.Error())
		return
	}

	// Import memory if present
	if len(bundle.Memory) > 0 {
		s.Memory.SetAll(ctx, task.ID, bundle.Memory)
	}

	writeJSON(w, http.StatusCreated, task)
}
