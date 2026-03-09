package model

import "time"

// --- Core entities ---

type Conversation struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Messages  []Message `json:"messages,omitempty"`
}

type Message struct {
	ID             string    `json:"id"`
	ConversationID string    `json:"conversation_id"`
	Role           string    `json:"role"` // user, assistant, system
	Content        string    `json:"content"`
	CreatedAt      time.Time `json:"created_at"`
}

type Task struct {
	ID             string       `json:"id"`
	Name           string       `json:"name"`
	Description    string       `json:"description"`
	PromptTemplate string       `json:"prompt_template"`
	InputSchema    []InputField `json:"input_schema"`
	OutputStyle    string       `json:"output_style"`
	Version        int          `json:"version"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
}

type InputField struct {
	Key     string   `json:"key"`
	Type    string   `json:"type"` // text, select, multi_select, number, boolean
	Label   string   `json:"label"`
	Options []string `json:"options,omitempty"`
	Default string   `json:"default,omitempty"`
}

type TaskVersion struct {
	ID        string    `json:"id"`
	TaskID    string    `json:"task_id"`
	Version   int       `json:"version"`
	Snapshot  string    `json:"snapshot"`
	CreatedAt time.Time `json:"created_at"`
}

type Run struct {
	ID           string                 `json:"id"`
	TaskID       string                 `json:"task_id"`
	TaskVersion  int                    `json:"task_version"`
	InputValues  map[string]interface{} `json:"input_values"`
	PromptFinal  string                 `json:"prompt_final"`
	Output       string                 `json:"output"`
	Status       string                 `json:"status"` // running, completed, failed
	Error        string                 `json:"error,omitempty"`
	Model        string                 `json:"model"`
	InputTokens  int                    `json:"input_tokens"`
	OutputTokens int                    `json:"output_tokens"`
	CostUSD      float64                `json:"cost_usd"`
	DurationMs   int64                  `json:"duration_ms"`
	CreatedAt    time.Time              `json:"created_at"`
}

type ProviderConfig struct {
	ID           string `json:"id"`
	Provider     string `json:"provider"`
	APIKey       string `json:"api_key,omitempty"`
	BaseURL      string `json:"base_url,omitempty"`
	DefaultModel string `json:"default_model,omitempty"`
	Configured   bool   `json:"configured"`
}

type MemoryEntry struct {
	TaskID    string    `json:"task_id"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

// --- Request types ---

type CreateMessageRequest struct {
	Content string `json:"content"`
	Model   string `json:"model,omitempty"`
}

type CreateTaskRequest struct {
	ConversationID string       `json:"conversation_id,omitempty"`
	Name           string       `json:"name"`
	Description    string       `json:"description,omitempty"`
	PromptTemplate string       `json:"prompt_template"`
	InputSchema    []InputField `json:"input_schema,omitempty"`
	OutputStyle    string       `json:"output_style,omitempty"`
}

type UpdateTaskRequest struct {
	Name           *string      `json:"name,omitempty"`
	Description    *string      `json:"description,omitempty"`
	PromptTemplate *string      `json:"prompt_template,omitempty"`
	InputSchema    []InputField `json:"input_schema,omitempty"`
	OutputStyle    *string      `json:"output_style,omitempty"`
}

type RunTaskRequest struct {
	Inputs map[string]interface{} `json:"inputs"`
	Model  string                 `json:"model,omitempty"`
}

type SaveProviderRequest struct {
	APIKey       string `json:"api_key"`
	BaseURL      string `json:"base_url,omitempty"`
	DefaultModel string `json:"default_model,omitempty"`
}

type SetMemoryRequest struct {
	Entries map[string]string `json:"entries,omitempty"`
	Value   string            `json:"value,omitempty"`
}

// --- Response / event types ---

type TaskDraft struct {
	Name           string       `json:"name"`
	Description    string       `json:"description"`
	PromptTemplate string       `json:"prompt_template"`
	InputSchema    []InputField `json:"input_schema"`
}

type ExportBundle struct {
	BundleVersion string            `json:"bundle_version"`
	Task          Task              `json:"task"`
	Versions      []TaskVersion     `json:"versions,omitempty"`
	Memory        map[string]string `json:"memory,omitempty"`
}

type StreamChunkEvent struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

type StreamDoneEvent struct {
	Type      string   `json:"type"`
	MessageID string   `json:"message_id"`
	Cost      CostInfo `json:"cost"`
}

type CostInfo struct {
	PromptTokens     int `json:"prompt"`
	CompletionTokens int `json:"completion"`
	TotalTokens      int `json:"total"`
}

// --- Objectives ---

type Objective struct {
	ID              string    `json:"id"`
	ConversationID  string    `json:"conversation_id"`
	Title           string    `json:"title"`
	Goal            string    `json:"goal"`
	Constraints     string    `json:"constraints"`
	ToolsRequired   []string  `json:"tools_required"`
	SuccessCriteria string    `json:"success_criteria"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type UpsertObjectiveRequest struct {
	Title           string   `json:"title"`
	Goal            string   `json:"goal"`
	Constraints     string   `json:"constraints,omitempty"`
	ToolsRequired   []string `json:"tools_required,omitempty"`
	SuccessCriteria string   `json:"success_criteria,omitempty"`
}

// --- Commands (Load 6) ---

type CommandExecuteRequest struct {
	Input   string `json:"input"`
	Confirm bool   `json:"confirm"`
}

type CommandExecuteResponse struct {
	OK         bool   `json:"ok"`
	Output     string `json:"output"`
	Stderr     string `json:"stderr,omitempty"`
	ExitCode   int    `json:"exit_code"`
	DurationMs int64  `json:"duration_ms"`
}

// --- Agents (Load 6, 8) ---

type Agent struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Slug         string            `json:"slug"`
	Description  string            `json:"description"`
	Instructions string            `json:"instructions"`
	ScriptBundle AgentScriptBundle `json:"script_bundle,omitempty"`
	Enabled      bool              `json:"enabled"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

type AgentRunRequest struct {
	InputPayload map[string]any `json:"input_payload,omitempty"`
}

type AgentRunResponse struct {
	OK         bool   `json:"ok"`
	Output     string `json:"output"`
	ExitCode   int    `json:"exit_code"`
	DurationMs int64  `json:"duration_ms"`
}

type AgentScriptBundle struct {
	Type         string `json:"type"`
	SystemPrompt string `json:"system_prompt,omitempty"`
	Model        string `json:"model,omitempty"`
}

// --- MCP servers ---

type MCPServer struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Transport    string            `json:"transport"`
	CommandOrURL string            `json:"command_or_url"`
	Args         []string          `json:"args"`
	Env          map[string]string `json:"env"`
	Enabled      bool              `json:"enabled"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

type CreateMCPServerRequest struct {
	Name         string            `json:"name"`
	Transport    string            `json:"transport"`
	CommandOrURL string            `json:"command_or_url"`
	Args         []string          `json:"args,omitempty"`
	Env          map[string]string `json:"env,omitempty"`
	Enabled      *bool             `json:"enabled,omitempty"`
}

type UpdateMCPServerRequest struct {
	Name         *string           `json:"name,omitempty"`
	Transport    *string           `json:"transport,omitempty"`
	CommandOrURL *string           `json:"command_or_url,omitempty"`
	Args         []string          `json:"args,omitempty"`
	Env          map[string]string `json:"env,omitempty"`
	Enabled      *bool             `json:"enabled,omitempty"`
}

type MCPServerTestResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}
