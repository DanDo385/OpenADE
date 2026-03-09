package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"openade/internal/db"
	"openade/internal/llm"
	"openade/internal/model"
)

type TaskService struct {
	DB *sql.DB
}

func NewTaskService(database *sql.DB) *TaskService {
	return &TaskService{DB: database}
}

func (s *TaskService) Create(ctx context.Context, req model.CreateTaskRequest) (*model.Task, error) {
	id := uuid.NewString()
	now := db.FormatTime(time.Now())

	inputSchema := req.InputSchema
	if inputSchema == nil {
		inputSchema = []model.InputField{}
	}
	schemaJSON, err := json.Marshal(inputSchema)
	if err != nil {
		return nil, fmt.Errorf("marshaling input schema: %w", err)
	}

	outputStyle := req.OutputStyle
	if outputStyle == "" {
		outputStyle = "markdown"
	}

	_, err = s.DB.ExecContext(ctx,
		`INSERT INTO tasks (id, name, description, prompt_template, input_schema_json, output_style, version, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, 1, ?, ?)`,
		id, req.Name, req.Description, req.PromptTemplate, string(schemaJSON), outputStyle, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("creating task: %w", err)
	}

	// Store initial version snapshot
	s.saveVersionSnapshot(ctx, id, 1, now)

	return s.Get(ctx, id)
}

func (s *TaskService) List(ctx context.Context, query string) ([]model.Task, error) {
	var rows *sql.Rows
	var err error
	if query != "" {
		rows, err = s.DB.QueryContext(ctx,
			`SELECT id, name, description, prompt_template, input_schema_json, output_style, version, created_at, updated_at
			 FROM tasks WHERE name LIKE ? OR description LIKE ? ORDER BY updated_at DESC`,
			"%"+query+"%", "%"+query+"%",
		)
	} else {
		rows, err = s.DB.QueryContext(ctx,
			`SELECT id, name, description, prompt_template, input_schema_json, output_style, version, created_at, updated_at
			 FROM tasks ORDER BY updated_at DESC`,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("listing tasks: %w", err)
	}
	defer rows.Close()

	var tasks []model.Task
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, *t)
	}
	return tasks, rows.Err()
}

func (s *TaskService) Get(ctx context.Context, id string) (*model.Task, error) {
	row := s.DB.QueryRowContext(ctx,
		`SELECT id, name, description, prompt_template, input_schema_json, output_style, version, created_at, updated_at
		 FROM tasks WHERE id = ?`, id,
	)
	return scanTaskRow(row)
}

func (s *TaskService) Update(ctx context.Context, id string, req model.UpdateTaskRequest) (*model.Task, error) {
	existing, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("task not found")
	}

	// Apply partial updates
	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.Description != nil {
		existing.Description = *req.Description
	}
	if req.PromptTemplate != nil {
		existing.PromptTemplate = *req.PromptTemplate
	}
	if req.InputSchema != nil {
		existing.InputSchema = req.InputSchema
	}
	if req.OutputStyle != nil {
		existing.OutputStyle = *req.OutputStyle
	}

	newVersion := existing.Version + 1
	now := db.FormatTime(time.Now())

	schemaJSON, _ := json.Marshal(existing.InputSchema)

	_, err = s.DB.ExecContext(ctx,
		`UPDATE tasks SET name=?, description=?, prompt_template=?, input_schema_json=?, output_style=?, version=?, updated_at=? WHERE id=?`,
		existing.Name, existing.Description, existing.PromptTemplate, string(schemaJSON),
		existing.OutputStyle, newVersion, now, id,
	)
	if err != nil {
		return nil, fmt.Errorf("updating task: %w", err)
	}

	// Store version snapshot
	s.saveVersionSnapshot(ctx, id, newVersion, now)

	return s.Get(ctx, id)
}

func (s *TaskService) Delete(ctx context.Context, id string) error {
	res, err := s.DB.ExecContext(ctx, `DELETE FROM tasks WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("deleting task: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("task not found")
	}
	return nil
}

// DraftFromConversation uses a meta-LLM call to generate a task draft from
// conversation messages.
func (s *TaskService) DraftFromConversation(ctx context.Context, messages []model.Message, adapter llm.Adapter) (*model.TaskDraft, error) {
	systemPrompt := `You are a task template extractor. Given a conversation between a user and an assistant, extract a reusable task template.

Return ONLY a valid JSON object with these fields:
- name: A short name for the task (2-5 words)
- description: A one-sentence description of what the task does
- prompt_template: The user's request rewritten as a template with {{variable_name}} placeholders for parts that should be customizable
- inputs: An array of input field definitions, each with:
  - key: The variable name (matching the template)
  - type: One of "text", "select", "multi_select", "number", "boolean"
  - label: A human-readable label for the input

Only return valid JSON. Do not include any other text, markdown formatting, or code fences.`

	// Build the conversation transcript for the meta-LLM
	transcript := ""
	for _, m := range messages {
		transcript += fmt.Sprintf("[%s]: %s\n\n", m.Role, m.Content)
	}

	llmMessages := []llm.ChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: "Here is the conversation:\n\n" + transcript},
	}

	result, err := adapter.Complete(ctx, llmMessages, "")
	if err != nil {
		return nil, fmt.Errorf("meta-LLM call failed: %w", err)
	}

	// Parse the LLM response as JSON
	var draft model.TaskDraft
	if err := json.Unmarshal([]byte(result.Content), &draft); err != nil {
		// Try to clean common issues (markdown code fences)
		cleaned := cleanJSONResponse(result.Content)
		if err2 := json.Unmarshal([]byte(cleaned), &draft); err2 != nil {
			return nil, fmt.Errorf("failed to parse meta-LLM response as JSON: %w (raw: %s)", err, result.Content)
		}
	}

	return &draft, nil
}

// Export creates an export bundle for a task.
func (s *TaskService) Export(ctx context.Context, id string) (*model.ExportBundle, error) {
	task, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, fmt.Errorf("task not found")
	}

	versions, err := s.getVersions(ctx, id)
	if err != nil {
		return nil, err
	}

	return &model.ExportBundle{
		BundleVersion: "0.1",
		Task:          *task,
		Versions:      versions,
	}, nil
}

// Import imports a task from an export bundle.
func (s *TaskService) Import(ctx context.Context, bundle model.ExportBundle) (*model.Task, error) {
	req := model.CreateTaskRequest{
		Name:           bundle.Task.Name,
		Description:    bundle.Task.Description,
		PromptTemplate: bundle.Task.PromptTemplate,
		InputSchema:    bundle.Task.InputSchema,
		OutputStyle:    bundle.Task.OutputStyle,
	}
	return s.Create(ctx, req)
}

func (s *TaskService) saveVersionSnapshot(ctx context.Context, taskID string, version int, now string) {
	task, err := s.Get(ctx, taskID)
	if err != nil || task == nil {
		return
	}
	snapshot, _ := json.Marshal(task)
	versionID := uuid.NewString()
	s.DB.ExecContext(ctx,
		`INSERT INTO task_versions (id, task_id, version, snapshot, created_at) VALUES (?, ?, ?, ?, ?)`,
		versionID, taskID, version, string(snapshot), now,
	)
}

func (s *TaskService) getVersions(ctx context.Context, taskID string) ([]model.TaskVersion, error) {
	rows, err := s.DB.QueryContext(ctx,
		`SELECT id, task_id, version, snapshot, created_at FROM task_versions WHERE task_id = ? ORDER BY version DESC`,
		taskID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing task versions: %w", err)
	}
	defer rows.Close()

	var versions []model.TaskVersion
	for rows.Next() {
		var v model.TaskVersion
		var createdAt string
		if err := rows.Scan(&v.ID, &v.TaskID, &v.Version, &v.Snapshot, &createdAt); err != nil {
			return nil, fmt.Errorf("scanning task version: %w", err)
		}
		v.CreatedAt = db.ParseTime(createdAt)
		versions = append(versions, v)
	}
	return versions, rows.Err()
}

// --- helpers ---

type rowScanner interface {
	Scan(dest ...interface{}) error
}

func scanTaskFromRow(scanner rowScanner) (*model.Task, error) {
	var t model.Task
	var schemaJSON, createdAt, updatedAt string
	err := scanner.Scan(&t.ID, &t.Name, &t.Description, &t.PromptTemplate,
		&schemaJSON, &t.OutputStyle, &t.Version, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scanning task: %w", err)
	}
	t.CreatedAt = db.ParseTime(createdAt)
	t.UpdatedAt = db.ParseTime(updatedAt)
	json.Unmarshal([]byte(schemaJSON), &t.InputSchema)
	if t.InputSchema == nil {
		t.InputSchema = []model.InputField{}
	}
	return &t, nil
}

func scanTask(rows *sql.Rows) (*model.Task, error) {
	return scanTaskFromRow(rows)
}

func scanTaskRow(row *sql.Row) (*model.Task, error) {
	return scanTaskFromRow(row)
}

func cleanJSONResponse(s string) string {
	// Strip common markdown code fences wrapping JSON
	s = trimPrefix(s, "```json\n")
	s = trimPrefix(s, "```json")
	s = trimPrefix(s, "```\n")
	s = trimPrefix(s, "```")
	s = trimSuffix(s, "\n```")
	s = trimSuffix(s, "```")
	return s
}

func trimPrefix(s, prefix string) string {
	if len(s) >= len(prefix) && s[:len(prefix)] == prefix {
		return s[len(prefix):]
	}
	return s
}

func trimSuffix(s, suffix string) string {
	if len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix {
		return s[:len(s)-len(suffix)]
	}
	return s
}
