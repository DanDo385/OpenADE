package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"openade/internal/llm"
	"openade/internal/model"
)

func (s *Server) HandleCreateConversation(w http.ResponseWriter, r *http.Request) {
	conv, err := s.Conversations.Create(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "create_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, conv)
}

func (s *Server) HandleListConversations(w http.ResponseWriter, r *http.Request) {
	convs, err := s.Conversations.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list_failed", err.Error())
		return
	}
	if convs == nil {
		convs = []model.Conversation{}
	}
	writeJSON(w, http.StatusOK, convs)
}

func (s *Server) HandleGetConversation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	conv, err := s.Conversations.Get(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "get_failed", err.Error())
		return
	}
	if conv == nil {
		writeError(w, http.StatusNotFound, "not_found", "conversation not found")
		return
	}
	writeJSON(w, http.StatusOK, conv)
}

func (s *Server) HandleDeleteConversation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := s.Conversations.Delete(r.Context(), id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "not_found", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "delete_failed", err.Error())
		return
	}
	writeOK(w)
}

// HandlePostMessage handles POST /api/conversations/:id/messages
// It persists the user message, streams the LLM response via SSE, and
// persists the assistant message on completion.
func (s *Server) HandlePostMessage(w http.ResponseWriter, r *http.Request) {
	convID := chi.URLParam(r, "id")

	var req model.CreateMessageRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "invalid request body")
		return
	}
	if strings.TrimSpace(req.Content) == "" {
		writeError(w, http.StatusBadRequest, "empty_content", "message content is required")
		return
	}

	ctx := r.Context()

	// Verify conversation exists
	conv, err := s.Conversations.Get(ctx, convID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "get_failed", err.Error())
		return
	}
	if conv == nil {
		writeError(w, http.StatusNotFound, "not_found", "conversation not found")
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

	// Persist user message
	_, err = s.Conversations.AddMessage(ctx, convID, "user", req.Content)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "save_failed", err.Error())
		return
	}

	// Build message history for LLM
	messages, err := s.Conversations.GetMessages(ctx, convID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "get_messages_failed", err.Error())
		return
	}
	llmMessages := make([]llm.ChatMessage, len(messages))
	for i, m := range messages {
		llmMessages[i] = llm.ChatMessage{Role: m.Role, Content: m.Content}
	}

	// Resolve model
	chatModel := req.Model
	if chatModel == "" && provCfg != nil {
		chatModel = provCfg.DefaultModel
	}

	// Start streaming
	stream, err := adapter.Stream(ctx, llmMessages, chatModel)
	if err != nil {
		writeError(w, http.StatusBadGateway, "stream_failed", err.Error())
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "streaming_unsupported", "streaming not supported")
		return
	}

	var fullContent strings.Builder
	var finalUsage *llm.Usage

	for event := range stream {
		if event.Error != nil {
			errEvent := map[string]string{"type": "error", "message": event.Error.Error()}
			data, _ := json.Marshal(errEvent)
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
			break
		}

		if event.Done {
			finalUsage = event.Usage
			break
		}

		// Send chunk event
		fullContent.WriteString(event.Content)
		chunk := model.StreamChunkEvent{Type: "chunk", Content: event.Content}
		data, _ := json.Marshal(chunk)
		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()
	}

	// Persist assistant message
	assistantMsg, err := s.Conversations.AddMessage(ctx, convID, "assistant", fullContent.String())
	if err != nil {
		log.Printf("ERROR: failed to persist assistant message: %v", err)
	}

	// Send done event
	msgID := ""
	if assistantMsg != nil {
		msgID = assistantMsg.ID
	}

	costInfo := model.CostInfo{}
	if finalUsage != nil {
		costInfo = model.CostInfo{
			PromptTokens:     finalUsage.PromptTokens,
			CompletionTokens: finalUsage.CompletionTokens,
			TotalTokens:      finalUsage.TotalTokens,
		}
	}

	doneEvent := model.StreamDoneEvent{
		Type:      "done",
		MessageID: msgID,
		Cost:      costInfo,
	}
	data, _ := json.Marshal(doneEvent)
	fmt.Fprintf(w, "data: %s\n\n", data)
	flusher.Flush()
}

// HandleDraftTaskFromConversation generates a task draft from conversation
// messages using a meta-LLM call.
func (s *Server) HandleDraftTaskFromConversation(w http.ResponseWriter, r *http.Request) {
	convID := chi.URLParam(r, "id")
	ctx := r.Context()

	conv, err := s.Conversations.Get(ctx, convID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "get_failed", err.Error())
		return
	}
	if conv == nil {
		writeError(w, http.StatusNotFound, "not_found", "conversation not found")
		return
	}
	if len(conv.Messages) == 0 {
		writeError(w, http.StatusBadRequest, "no_messages", "conversation has no messages")
		return
	}

	adapter, _, err := s.getLLMAdapter(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "provider_error", err.Error())
		return
	}
	if adapter == nil {
		writeError(w, http.StatusUnauthorized, "no_provider", "no LLM provider configured")
		return
	}

	draft, err := s.Tasks.DraftFromConversation(ctx, conv.Messages, adapter)
	if err != nil {
		writeError(w, http.StatusBadGateway, "draft_failed", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, draft)
}

