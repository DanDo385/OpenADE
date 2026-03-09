package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"os/exec"
	"strings"
	"time"

	"github.com/google/uuid"
	"openade/internal/db"
	"openade/internal/model"
)

type MCPServerService struct {
	DB *sql.DB
}

func NewMCPServerService(database *sql.DB) *MCPServerService {
	return &MCPServerService{DB: database}
}

func (s *MCPServerService) List(ctx context.Context) ([]model.MCPServer, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT id, name, transport, command_or_url, args_json, env_json, enabled, created_at, updated_at
		FROM mcp_servers
		ORDER BY name
	`)
	if err != nil {
		return nil, fmt.Errorf("listing mcp servers: %w", err)
	}
	defer rows.Close()

	var servers []model.MCPServer
	for rows.Next() {
		server, err := scanMCPServer(rows)
		if err != nil {
			return nil, err
		}
		servers = append(servers, *server)
	}
	return servers, rows.Err()
}

func (s *MCPServerService) Get(ctx context.Context, id string) (*model.MCPServer, error) {
	row := s.DB.QueryRowContext(ctx, `
		SELECT id, name, transport, command_or_url, args_json, env_json, enabled, created_at, updated_at
		FROM mcp_servers
		WHERE id = ?
	`, id)
	return scanMCPServerRow(row)
}

func (s *MCPServerService) Create(ctx context.Context, req model.CreateMCPServerRequest) (*model.MCPServer, error) {
	payload, err := normalizeMCPServerCreate(req)
	if err != nil {
		return nil, err
	}

	id := uuid.NewString()
	now := db.FormatTime(time.Now())
	argsJSON, _ := json.Marshal(payload.Args)
	envJSON, _ := json.Marshal(payload.Env)

	_, err = s.DB.ExecContext(ctx, `
		INSERT INTO mcp_servers (id, name, transport, command_or_url, args_json, env_json, enabled, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, id, payload.Name, payload.Transport, payload.CommandOrURL, string(argsJSON), string(envJSON), boolToInt(payload.Enabled), now, now)
	if err != nil {
		return nil, fmt.Errorf("creating mcp server: %w", err)
	}

	return s.Get(ctx, id)
}

func (s *MCPServerService) Update(ctx context.Context, id string, req model.UpdateMCPServerRequest) (*model.MCPServer, error) {
	current, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, fmt.Errorf("mcp server not found")
	}

	payload, err := normalizeMCPServerUpdate(*current, req)
	if err != nil {
		return nil, err
	}

	argsJSON, _ := json.Marshal(payload.Args)
	envJSON, _ := json.Marshal(payload.Env)
	now := db.FormatTime(time.Now())

	_, err = s.DB.ExecContext(ctx, `
		UPDATE mcp_servers
		SET name = ?, transport = ?, command_or_url = ?, args_json = ?, env_json = ?, enabled = ?, updated_at = ?
		WHERE id = ?
	`, payload.Name, payload.Transport, payload.CommandOrURL, string(argsJSON), string(envJSON), boolToInt(payload.Enabled), now, id)
	if err != nil {
		return nil, fmt.Errorf("updating mcp server: %w", err)
	}

	return s.Get(ctx, id)
}

func (s *MCPServerService) Delete(ctx context.Context, id string) error {
	res, err := s.DB.ExecContext(ctx, `DELETE FROM mcp_servers WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("deleting mcp server: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("mcp server not found")
	}
	return nil
}

func (s *MCPServerService) Test(ctx context.Context, id string) (*model.MCPServerTestResponse, error) {
	server, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if server == nil {
		return nil, fmt.Errorf("mcp server not found")
	}

	switch server.Transport {
	case "stdio":
		if _, err := exec.LookPath(server.CommandOrURL); err != nil {
			return &model.MCPServerTestResponse{
				OK:      false,
				Message: fmt.Sprintf("command not found: %s", server.CommandOrURL),
			}, nil
		}
		return &model.MCPServerTestResponse{
			OK:      true,
			Message: fmt.Sprintf("command is available: %s", server.CommandOrURL),
		}, nil
	case "sse":
		if _, err := url.ParseRequestURI(server.CommandOrURL); err != nil {
			return &model.MCPServerTestResponse{
				OK:      false,
				Message: "invalid URL",
			}, nil
		}
		return &model.MCPServerTestResponse{
			OK:      true,
			Message: "URL is syntactically valid",
		}, nil
	default:
		return nil, fmt.Errorf("unsupported transport: %s", server.Transport)
	}
}

type mcpServerRowScanner interface {
	Scan(dest ...any) error
}

func scanMCPServer(rows *sql.Rows) (*model.MCPServer, error) {
	return scanMCPServerFromScanner(rows)
}

func scanMCPServerRow(row *sql.Row) (*model.MCPServer, error) {
	return scanMCPServerFromScanner(row)
}

func scanMCPServerFromScanner(scanner mcpServerRowScanner) (*model.MCPServer, error) {
	var server model.MCPServer
	var argsJSON, envJSON, createdAt, updatedAt string
	var enabled int
	err := scanner.Scan(
		&server.ID,
		&server.Name,
		&server.Transport,
		&server.CommandOrURL,
		&argsJSON,
		&envJSON,
		&enabled,
		&createdAt,
		&updatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scanning mcp server: %w", err)
	}

	server.Enabled = enabled != 0
	server.CreatedAt = db.ParseTime(createdAt)
	server.UpdatedAt = db.ParseTime(updatedAt)
	if err := json.Unmarshal([]byte(argsJSON), &server.Args); err != nil || server.Args == nil {
		server.Args = []string{}
	}
	if err := json.Unmarshal([]byte(envJSON), &server.Env); err != nil || server.Env == nil {
		server.Env = map[string]string{}
	}

	return &server, nil
}

func normalizeMCPServerCreate(req model.CreateMCPServerRequest) (*model.MCPServer, error) {
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	server := &model.MCPServer{
		Name:         req.Name,
		Transport:    req.Transport,
		CommandOrURL: req.CommandOrURL,
		Args:         req.Args,
		Env:          req.Env,
		Enabled:      enabled,
	}
	return normalizeMCPServer(server)
}

func normalizeMCPServerUpdate(current model.MCPServer, req model.UpdateMCPServerRequest) (*model.MCPServer, error) {
	if req.Name != nil {
		current.Name = *req.Name
	}
	if req.Transport != nil {
		current.Transport = *req.Transport
	}
	if req.CommandOrURL != nil {
		current.CommandOrURL = *req.CommandOrURL
	}
	if req.Args != nil {
		current.Args = req.Args
	}
	if req.Env != nil {
		current.Env = req.Env
	}
	if req.Enabled != nil {
		current.Enabled = *req.Enabled
	}
	return normalizeMCPServer(&current)
}

func normalizeMCPServer(server *model.MCPServer) (*model.MCPServer, error) {
	server.Name = strings.TrimSpace(server.Name)
	server.Transport = strings.ToLower(strings.TrimSpace(server.Transport))
	server.CommandOrURL = strings.TrimSpace(server.CommandOrURL)
	if server.Args == nil {
		server.Args = []string{}
	}
	if server.Env == nil {
		server.Env = map[string]string{}
	}

	if server.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if server.CommandOrURL == "" {
		return nil, fmt.Errorf("command_or_url is required")
	}
	switch server.Transport {
	case "stdio", "sse":
	default:
		return nil, fmt.Errorf("transport must be one of: stdio, sse")
	}

	return server, nil
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
