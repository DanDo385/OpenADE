package handlers

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"openade/internal/model"
)

func (s *Server) HandleListSchedules(w http.ResponseWriter, r *http.Request) {
	taskID := r.URL.Query().Get("task_id")
	schedules, err := s.Schedules.List(r.Context(), taskID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list_failed", err.Error())
		return
	}
	if schedules == nil {
		schedules = []model.Schedule{}
	}
	writeJSON(w, http.StatusOK, schedules)
}

func (s *Server) HandleCreateSchedule(w http.ResponseWriter, r *http.Request) {
	var req model.CreateScheduleRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "invalid request body")
		return
	}

	schedule, err := s.Schedules.Create(r.Context(), req)
	if err != nil {
		status := http.StatusBadRequest
		if strings.Contains(err.Error(), "creating schedule") {
			status = http.StatusInternalServerError
		}
		writeError(w, status, "create_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, schedule)
}

func (s *Server) HandleUpdateSchedule(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req model.UpdateScheduleRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "invalid request body")
		return
	}

	schedule, err := s.Schedules.Update(r.Context(), id, req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "not_found", err.Error())
			return
		}
		writeError(w, http.StatusBadRequest, "update_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, schedule)
}

func (s *Server) HandleDeleteSchedule(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := s.Schedules.Delete(r.Context(), id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "not_found", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "delete_failed", err.Error())
		return
	}
	writeOK(w)
}

func (s *Server) HandleGetTaskSchedule(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")
	schedule, err := s.Schedules.GetByTaskID(r.Context(), taskID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "get_failed", err.Error())
		return
	}
	if schedule == nil {
		writeError(w, http.StatusNotFound, "not_found", "schedule not found")
		return
	}
	writeJSON(w, http.StatusOK, schedule)
}

func (s *Server) HandleUpsertTaskSchedule(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")
	var req model.UpdateScheduleRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "invalid request body")
		return
	}

	schedule, err := s.Schedules.UpsertForTask(r.Context(), taskID, req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "not_found", err.Error())
			return
		}
		writeError(w, http.StatusBadRequest, "upsert_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, schedule)
}

func (s *Server) HandleDeleteTaskSchedule(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")
	if err := s.Schedules.DeleteByTaskID(r.Context(), taskID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "not_found", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "delete_failed", err.Error())
		return
	}
	writeOK(w)
}
