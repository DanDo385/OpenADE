package handlers

import (
	"context"

	"openade/internal/llm"
	mcpclient "openade/internal/mcp"
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
	Commands      *services.CommandService
	Agents        *services.AgentService
	Objectives    *services.ObjectiveService
	MCPServers    *services.MCPServerService
	MCPClients    *mcpclient.ClientManager
	Schedules     *services.SchedulerService
}

// NewServer creates a Server with all services wired up.
func NewServer(
	convSvc *services.ConversationService,
	taskSvc *services.TaskService,
	runSvc *services.RunService,
	memSvc *services.MemoryService,
	provSvc *services.ProviderService,
	cmdSvc *services.CommandService,
	agentSvc *services.AgentService,
	objSvc *services.ObjectiveService,
	mcpSvc *services.MCPServerService,
	mcpClientMgr *mcpclient.ClientManager,
	scheduleSvc *services.SchedulerService,
) *Server {
	return &Server{
		Conversations: convSvc,
		Tasks:         taskSvc,
		Runs:          runSvc,
		Memory:        memSvc,
		Providers:     provSvc,
		Commands:      cmdSvc,
		Agents:        agentSvc,
		Objectives:    objSvc,
		MCPServers:    mcpSvc,
		MCPClients:    mcpClientMgr,
		Schedules:     scheduleSvc,
	}
}

// RegisterRoutes mounts all API routes onto the given chi router.
func (s *Server) RegisterRoutes(r chi.Router) {
	r.Get("/", s.HandleRoot)
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
				r.Get("/objective", s.HandleGetObjective)
				r.Put("/objective", s.HandleUpsertObjective)
				r.Get("/objective/export", s.HandleExportObjectiveMarkdown)
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
				r.Get("/schedule", s.HandleGetTaskSchedule)
				r.Put("/schedule", s.HandleUpsertTaskSchedule)
				r.Delete("/schedule", s.HandleDeleteTaskSchedule)
			})
		})

		// Schedules
		r.Route("/schedules", func(r chi.Router) {
			r.Get("/", s.HandleListSchedules)
			r.Post("/", s.HandleCreateSchedule)
			r.Route("/{id}", func(r chi.Router) {
				r.Put("/", s.HandleUpdateSchedule)
				r.Delete("/", s.HandleDeleteSchedule)
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

		// Commands (Load 6)
		r.Post("/commands/execute", s.HandleExecuteCommand)

		// Agents (Load 6, 8)
		r.Get("/agents", s.HandleListAgents)
		r.Get("/agents/slug/{slug}", s.HandleGetAgentBySlug)
		r.Route("/agents/{id}", func(r chi.Router) {
			r.Get("/", s.HandleGetAgent)
			r.Post("/run", s.HandleRunAgent)
		})

		// MCP servers (Phase 2 foundation)
		r.Route("/mcp/servers", func(r chi.Router) {
			r.Get("/", s.HandleListMCPServers)
			r.Post("/", s.HandleCreateMCPServer)
			r.Route("/{id}", func(r chi.Router) {
				r.Put("/", s.HandleUpdateMCPServer)
				r.Delete("/", s.HandleDeleteMCPServer)
				r.Post("/test", s.HandleTestMCPServer)
				r.Get("/tools", s.HandleListMCPServerTools)
			})
		})
		r.Post("/mcp/tools/call", s.HandleCallMCPTool)

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
