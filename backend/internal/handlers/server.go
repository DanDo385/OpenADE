package handlers

import (
	"context"

	"openade/internal/llm"
	"openade/internal/model"
	"openade/internal/services"

	"github.com/go-chi/chi/v5"
)

// Server holds all service dependencies and exposes HTTP handler methods.
type Server struct {
	Conversations *services.ConversationService
	Tasks         *services.TaskService
	Runs          *services.RunService
	Memory        *services.MemoryService
	Providers     *services.ProviderService
}

// NewServer creates a Server with all services wired up.
func NewServer(
	convSvc *services.ConversationService,
	taskSvc *services.TaskService,
	runSvc *services.RunService,
	memSvc *services.MemoryService,
	provSvc *services.ProviderService,
) *Server {
	return &Server{
		Conversations: convSvc,
		Tasks:         taskSvc,
		Runs:          runSvc,
		Memory:        memSvc,
		Providers:     provSvc,
	}
}

// RegisterRoutes mounts all API routes onto the given chi router.
func (s *Server) RegisterRoutes(r chi.Router) {
	r.Get("/health", s.HandleHealth)

	r.Route("/api", func(r chi.Router) {
		// Conversations
		r.Route("/conversations", func(r chi.Router) {
			r.Post("/", s.HandleCreateConversation)
			r.Get("/", s.HandleListConversations)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", s.HandleGetConversation)
				r.Delete("/", s.HandleDeleteConversation)
				r.Post("/messages", s.HandlePostMessage)
				r.Post("/draft-task", s.HandleDraftTaskFromConversation)
			})
		})

		// Tasks
		r.Route("/tasks", func(r chi.Router) {
			r.Post("/", s.HandleCreateTask)
			r.Get("/", s.HandleListTasks)
			r.Post("/import", s.HandleImportTask)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", s.HandleGetTask)
				r.Put("/", s.HandleUpdateTask)
				r.Delete("/", s.HandleDeleteTask)
				r.Post("/run", s.HandleRunTask)
				r.Post("/export", s.HandleExportTask)
			})
		})

		// Runs
		r.Route("/runs", func(r chi.Router) {
			r.Get("/", s.HandleListRuns)
			r.Get("/{id}", s.HandleGetRun)
		})

		// Providers
		r.Route("/providers", func(r chi.Router) {
			r.Get("/", s.HandleListProviders)
			r.Put("/{id}", s.HandleSaveProvider)
		})

		// Memory
		r.Route("/memory/{task_id}", func(r chi.Router) {
			r.Get("/", s.HandleGetMemory)
			r.Put("/", s.HandleSetMemory)
			r.Put("/{key}", s.HandleSetMemoryKey)
		})
	})
}

// getLLMAdapter creates an LLM adapter from the current provider configuration.
func (s *Server) getLLMAdapter(ctx context.Context) (llm.Adapter, *model.ProviderConfig, error) {
	cfg, err := s.Providers.GetDefault(ctx)
	if err != nil {
		return nil, nil, err
	}
	if cfg == nil {
		return nil, nil, nil
	}
	adapter := llm.NewOpenAI(cfg.APIKey, cfg.BaseURL, cfg.DefaultModel)
	return adapter, cfg, nil
}
