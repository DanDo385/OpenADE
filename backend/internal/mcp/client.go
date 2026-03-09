package mcpclient

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"

	gomcpclient "github.com/mark3labs/mcp-go/client"
	gomcp "github.com/mark3labs/mcp-go/mcp"
	"openade/internal/model"
)

type ServerStore interface {
	Get(ctx context.Context, id string) (*model.MCPServer, error)
}

type ClientManager struct {
	servers  ServerStore
	mu       sync.Mutex
	sessions map[string]*session
}

type session struct {
	server *model.MCPServer
	client *gomcpclient.Client
}

func NewClientManager(servers ServerStore) *ClientManager {
	return &ClientManager{
		servers:  servers,
		sessions: map[string]*session{},
	}
}

func (m *ClientManager) ListTools(ctx context.Context, serverID string) ([]model.MCPToolInfo, error) {
	sess, err := m.ensureSession(ctx, serverID)
	if err != nil {
		return nil, err
	}

	result, err := sess.client.ListTools(ctx, gomcp.ListToolsRequest{})
	if err != nil {
		return nil, fmt.Errorf("listing tools: %w", err)
	}

	tools := make([]model.MCPToolInfo, 0, len(result.Tools))
	for _, tool := range result.Tools {
		tools = append(tools, toolInfoFromSDK(tool))
	}
	return tools, nil
}

func (m *ClientManager) CallTool(ctx context.Context, serverID, toolName string, args map[string]any) (*gomcp.CallToolResult, error) {
	sess, err := m.ensureSession(ctx, serverID)
	if err != nil {
		return nil, err
	}

	result, err := sess.client.CallTool(ctx, gomcp.CallToolRequest{
		Params: gomcp.CallToolParams{
			Name:      toolName,
			Arguments: args,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("calling tool %q: %w", toolName, err)
	}
	return result, nil
}

func (m *ClientManager) TestServer(ctx context.Context, serverID string) (*model.MCPServerTestResponse, error) {
	tools, err := m.ListTools(ctx, serverID)
	if err != nil {
		return &model.MCPServerTestResponse{
			OK:      false,
			Message: err.Error(),
		}, nil
	}

	return &model.MCPServerTestResponse{
		OK:        true,
		Message:   "connected and listed tools",
		ToolCount: len(tools),
		Tools:     tools,
	}, nil
}

func (m *ClientManager) Close() error {
	m.mu.Lock()
	sessions := make([]*session, 0, len(m.sessions))
	for id, sess := range m.sessions {
		sessions = append(sessions, sess)
		delete(m.sessions, id)
	}
	m.mu.Unlock()

	var firstErr error
	for _, sess := range sessions {
		if err := sess.client.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (m *ClientManager) ensureSession(ctx context.Context, serverID string) (*session, error) {
	server, err := m.servers.Get(ctx, serverID)
	if err != nil {
		return nil, fmt.Errorf("loading mcp server: %w", err)
	}
	if server == nil {
		return nil, fmt.Errorf("mcp server not found")
	}
	if !server.Enabled {
		return nil, fmt.Errorf("mcp server is disabled")
	}
	if server.Transport != "stdio" {
		return nil, fmt.Errorf("unsupported transport for client manager: %s", server.Transport)
	}

	m.mu.Lock()
	existing := m.sessions[serverID]
	m.mu.Unlock()
	if existing != nil {
		if sessionMatchesConfig(existing.server, server) {
			return existing, nil
		}
		if err := existing.client.Close(); err != nil {
			// stale session close failure should not block re-connect
		}
		m.mu.Lock()
		delete(m.sessions, serverID)
		m.mu.Unlock()
	}

	client, err := gomcpclient.NewStdioMCPClient(server.CommandOrURL, envList(server.Env), server.Args...)
	if err != nil {
		return nil, fmt.Errorf("starting stdio mcp client: %w", err)
	}

	initReq := gomcp.InitializeRequest{}
	initReq.Params.ProtocolVersion = gomcp.LATEST_PROTOCOL_VERSION
	initReq.Params.ClientInfo = gomcp.Implementation{
		Name:    "openade",
		Version: "0.1.0",
	}
	initReq.Params.Capabilities = gomcp.ClientCapabilities{}

	if _, err := client.Initialize(ctx, initReq); err != nil {
		client.Close()
		return nil, fmt.Errorf("initializing mcp session: %w", err)
	}

	created := &session{
		server: cloneServer(server),
		client: client,
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if existing := m.sessions[serverID]; existing != nil {
		if err := client.Close(); err != nil {
			// connection was superseded by another goroutine; ignore close errors
		}
		return existing, nil
	}
	m.sessions[serverID] = created
	return created, nil
}

func toolInfoFromSDK(tool gomcp.Tool) model.MCPToolInfo {
	return model.MCPToolInfo{
		Name:         tool.Name,
		Description:  tool.Description,
		InputSchema:  schemaJSON(tool.RawInputSchema, tool.InputSchema),
		OutputSchema: schemaJSON(tool.RawOutputSchema, tool.OutputSchema),
	}
}

func schemaJSON(raw json.RawMessage, v any) json.RawMessage {
	if len(raw) > 0 {
		return append(json.RawMessage(nil), raw...)
	}
	data, err := json.Marshal(v)
	if err != nil || string(data) == "null" || string(data) == "{}" {
		return nil
	}
	return data
}

func envList(values map[string]string) []string {
	if len(values) == 0 {
		return nil
	}
	out := make([]string, 0, len(values))
	for key, value := range values {
		out = append(out, fmt.Sprintf("%s=%s", key, value))
	}
	return out
}

func cloneServer(server *model.MCPServer) *model.MCPServer {
	if server == nil {
		return nil
	}
	copyServer := *server
	copyServer.Args = append([]string(nil), server.Args...)
	copyServer.Env = map[string]string{}
	for key, value := range server.Env {
		copyServer.Env[key] = value
	}
	return &copyServer
}

func sessionMatchesConfig(existing, current *model.MCPServer) bool {
	if existing == nil || current == nil {
		return false
	}
	return existing.Transport == current.Transport &&
		existing.CommandOrURL == current.CommandOrURL &&
		existing.Enabled == current.Enabled &&
		reflect.DeepEqual(existing.Args, current.Args) &&
		reflect.DeepEqual(existing.Env, current.Env)
}
