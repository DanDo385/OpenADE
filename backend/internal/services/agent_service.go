package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"openade/internal/db"
	"openade/internal/llm"
	"openade/internal/model"
)

type AgentService struct {
	DB         *sql.DB
	Providers  *ProviderService
	NewAdapter func(cfg *model.ProviderConfig) llm.Adapter
}

func NewAgentService(database *sql.DB, providers *ProviderService, newAdapter func(cfg *model.ProviderConfig) llm.Adapter) *AgentService {
	return &AgentService{
		DB:         database,
		Providers:  providers,
		NewAdapter: newAdapter,
	}
}

func (s *AgentService) List(ctx context.Context) ([]model.Agent, error) {
	rows, err := s.DB.QueryContext(ctx,
		`SELECT id, name, slug, description, instructions, script_bundle_json, enabled, created_at, updated_at
		 FROM agents ORDER BY name`,
	)
	if err != nil {
		return nil, fmt.Errorf("listing agents: %w", err)
	}
	defer rows.Close()

	var agents []model.Agent
	for rows.Next() {
		var a model.Agent
		var bundleJSON, createdAt, updatedAt string
		if err := rows.Scan(&a.ID, &a.Name, &a.Slug, &a.Description, &a.Instructions,
			&bundleJSON, &a.Enabled, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		json.Unmarshal([]byte(bundleJSON), &a.ScriptBundle)
		a.CreatedAt = db.ParseTime(createdAt)
		a.UpdatedAt = db.ParseTime(updatedAt)
		agents = append(agents, a)
	}
	return agents, rows.Err()
}

func (s *AgentService) GetByID(ctx context.Context, id string) (*model.Agent, error) {
	return s.getAgent(ctx, "id", id)
}

func (s *AgentService) GetBySlug(ctx context.Context, slug string) (*model.Agent, error) {
	return s.getAgent(ctx, "slug", slug)
}

func (s *AgentService) getAgent(ctx context.Context, col, val string) (*model.Agent, error) {
	query := fmt.Sprintf(`SELECT id, name, slug, description, instructions, script_bundle_json, enabled, created_at, updated_at
		FROM agents WHERE %s = ?`, col)
	var a model.Agent
	var bundleJSON, createdAt, updatedAt string
	err := s.DB.QueryRowContext(ctx, query, val).Scan(
		&a.ID, &a.Name, &a.Slug, &a.Description, &a.Instructions,
		&bundleJSON, &a.Enabled, &createdAt, &updatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting agent: %w", err)
	}
	json.Unmarshal([]byte(bundleJSON), &a.ScriptBundle)
	a.CreatedAt = db.ParseTime(createdAt)
	a.UpdatedAt = db.ParseTime(updatedAt)
	return &a, nil
}

func (s *AgentService) Run(ctx context.Context, id string, req model.AgentRunRequest) (*model.AgentRunResponse, error) {
	agent, err := s.GetByID(ctx, id)
	if err != nil || agent == nil {
		return nil, fmt.Errorf("agent not found")
	}
	if !agent.Enabled {
		return nil, fmt.Errorf("agent is disabled")
	}

	if s.Providers == nil || s.NewAdapter == nil {
		return nil, errors.New("agent service is missing LLM dependencies")
	}

	provCfg, err := s.Providers.GetDefault(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading provider config: %w", err)
	}
	if provCfg == nil {
		return nil, errors.New("no LLM provider configured")
	}

	bundle := agent.ScriptBundle
	if bundle.Type == "" {
		bundle.Type = "prompt"
	}
	if bundle.Type != "prompt" {
		return nil, fmt.Errorf("unsupported agent script bundle type: %s", bundle.Type)
	}

	userMessage := strings.TrimSpace(stringFromMap(req.InputPayload, "message"))
	if userMessage == "" {
		return nil, errors.New("input_payload.message is required")
	}

	systemPrompt := strings.TrimSpace(joinNonEmpty("\n\n", bundle.SystemPrompt, agent.Instructions))
	messages := make([]llm.ChatMessage, 0, 2)
	if systemPrompt != "" {
		messages = append(messages, llm.ChatMessage{Role: "system", Content: systemPrompt})
	}
	messages = append(messages, llm.ChatMessage{Role: "user", Content: userMessage})

	modelName := bundle.Model
	if modelName == "" {
		modelName = provCfg.DefaultModel
	}

	start := time.Now()
	result, err := s.NewAdapter(provCfg).Complete(ctx, messages, modelName)
	dur := time.Since(start).Milliseconds()
	if err != nil {
		return nil, fmt.Errorf("running agent with llm: %w", err)
	}

	return &model.AgentRunResponse{
		OK:         true,
		Output:     result.Content,
		ExitCode:   0,
		DurationMs: dur,
	}, nil
}

func (s *AgentService) Create(ctx context.Context, name, slug, description, instructions string, scriptBundle model.AgentScriptBundle) (*model.Agent, error) {
	if slug == "" {
		slug = stringsToSlug(name)
	}
	id := uuid.NewString()
	now := db.FormatTime(time.Now())
	bundleJSON, _ := json.Marshal(scriptBundle)
	if bundleJSON == nil {
		bundleJSON = []byte("{}")
	}

	_, err := s.DB.ExecContext(ctx,
		`INSERT INTO agents (id, name, slug, description, instructions, script_bundle_json, enabled, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, 1, ?, ?)`,
		id, name, slug, description, instructions, string(bundleJSON), now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("creating agent: %w", err)
	}
	return s.GetByID(ctx, id)
}

func stringsToSlug(s string) string {
	var b []byte
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b = append(b, byte(r))
		} else if r >= 'A' && r <= 'Z' {
			b = append(b, byte(r+32))
		} else if r == ' ' || r == '-' {
			b = append(b, '-')
		}
	}
	return string(b)
}

func stringFromMap(values map[string]any, key string) string {
	if values == nil {
		return ""
	}
	switch v := values[key].(type) {
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	case nil:
		return ""
	default:
		return fmt.Sprint(v)
	}
}

func joinNonEmpty(sep string, parts ...string) string {
	filtered := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			filtered = append(filtered, part)
		}
	}
	return strings.Join(filtered, sep)
}
