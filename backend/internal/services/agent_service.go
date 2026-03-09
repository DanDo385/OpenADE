package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"openade/internal/db"
	"openade/internal/model"
)

type AgentService struct {
	DB *sql.DB
}

func NewAgentService(database *sql.DB) *AgentService {
	return &AgentService{DB: database}
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

	start := time.Now()
	// For MVP: agents run as echo of instructions + input (no real script execution)
	// In Load 6 full implementation, script_bundle would define actual commands
	output := "Agent: " + agent.Name + "\n"
	if agent.Instructions != "" {
		output += "Instructions: " + agent.Instructions + "\n"
	}
	if len(req.InputPayload) > 0 {
		output += "Input: " + fmt.Sprintf("%v", req.InputPayload)
	}
	if output == "Agent: "+agent.Name+"\n" {
		output += "Run complete. (No script bundle configured yet.)"
	}

	dur := time.Since(start).Milliseconds()
	return &model.AgentRunResponse{
		OK:         true,
		Output:     output,
		ExitCode:   0,
		DurationMs: dur,
	}, nil
}

func (s *AgentService) Create(ctx context.Context, name, slug, description, instructions string, scriptBundle map[string]any) (*model.Agent, error) {
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
