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

type RunService struct {
	DB *sql.DB
}

func NewRunService(database *sql.DB) *RunService {
	return &RunService{DB: database}
}

// Execute runs a task with the given inputs. It renders the prompt template,
// calls the LLM (non-streaming), and persists the run record.
func (s *RunService) Execute(ctx context.Context, task *model.Task, inputs map[string]interface{}, adapter llm.Adapter, modelOverride string) (*model.Run, error) {
	// Render prompt template
	promptFinal, err := RenderTemplate(task.PromptTemplate, inputs)
	if err != nil {
		return nil, fmt.Errorf("rendering template: %w", err)
	}

	// Create run record with "running" status
	runID := uuid.NewString()
	now := db.FormatTime(time.Now())
	inputsJSON, _ := json.Marshal(inputs)

	_, err = s.DB.ExecContext(ctx,
		`INSERT INTO runs (id, task_id, task_version, input_values_json, prompt_final, status, model, created_at)
		 VALUES (?, ?, ?, ?, ?, 'running', ?, ?)`,
		runID, task.ID, task.Version, string(inputsJSON), promptFinal, modelOverride, now,
	)
	if err != nil {
		return nil, fmt.Errorf("creating run record: %w", err)
	}

	// Call LLM (non-streaming for task runs)
	messages := []llm.ChatMessage{
		{Role: "user", Content: promptFinal},
	}

	startTime := time.Now()
	result, llmErr := adapter.Complete(ctx, messages, modelOverride)
	durationMs := time.Since(startTime).Milliseconds()

	if llmErr != nil {
		// Mark run as failed
		s.DB.ExecContext(ctx,
			`UPDATE runs SET status='failed', error_text=?, duration_ms=? WHERE id=?`,
			llmErr.Error(), durationMs, runID,
		)
		return nil, fmt.Errorf("LLM execution failed: %w", llmErr)
	}

	// Calculate cost
	costUSD := llm.EstimateCost(modelOverride, result.Usage.PromptTokens, result.Usage.CompletionTokens)

	// Update run with results
	_, err = s.DB.ExecContext(ctx,
		`UPDATE runs SET output=?, status='completed', model=?, input_tokens=?, output_tokens=?, cost_usd=?, duration_ms=? WHERE id=?`,
		result.Content, modelOverride, result.Usage.PromptTokens, result.Usage.CompletionTokens,
		costUSD, durationMs, runID,
	)
	if err != nil {
		return nil, fmt.Errorf("updating run: %w", err)
	}

	return s.Get(ctx, runID)
}

func (s *RunService) List(ctx context.Context, taskID string) ([]model.Run, error) {
	var rows *sql.Rows
	var err error
	if taskID != "" {
		rows, err = s.DB.QueryContext(ctx,
			`SELECT id, task_id, task_version, input_values_json, prompt_final, output, status, error_text,
			        model, input_tokens, output_tokens, cost_usd, duration_ms, created_at
			 FROM runs WHERE task_id = ? ORDER BY created_at DESC`,
			taskID,
		)
	} else {
		rows, err = s.DB.QueryContext(ctx,
			`SELECT id, task_id, task_version, input_values_json, prompt_final, output, status, error_text,
			        model, input_tokens, output_tokens, cost_usd, duration_ms, created_at
			 FROM runs ORDER BY created_at DESC`,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("listing runs: %w", err)
	}
	defer rows.Close()

	var runs []model.Run
	for rows.Next() {
		r, err := scanRun(rows)
		if err != nil {
			return nil, err
		}
		runs = append(runs, *r)
	}
	return runs, rows.Err()
}

func (s *RunService) Get(ctx context.Context, id string) (*model.Run, error) {
	row := s.DB.QueryRowContext(ctx,
		`SELECT id, task_id, task_version, input_values_json, prompt_final, output, status, error_text,
		        model, input_tokens, output_tokens, cost_usd, duration_ms, created_at
		 FROM runs WHERE id = ?`, id,
	)
	return scanRunRow(row)
}

func scanRun(rows *sql.Rows) (*model.Run, error) {
	return scanRunFromScanner(rows)
}

func scanRunRow(row *sql.Row) (*model.Run, error) {
	return scanRunFromScanner(row)
}

func scanRunFromScanner(scanner rowScanner) (*model.Run, error) {
	var r model.Run
	var inputsJSON, createdAt, errorText string
	err := scanner.Scan(
		&r.ID, &r.TaskID, &r.TaskVersion, &inputsJSON, &r.PromptFinal,
		&r.Output, &r.Status, &errorText, &r.Model,
		&r.InputTokens, &r.OutputTokens, &r.CostUSD, &r.DurationMs, &createdAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scanning run: %w", err)
	}
	r.CreatedAt = db.ParseTime(createdAt)
	r.Error = errorText
	json.Unmarshal([]byte(inputsJSON), &r.InputValues)
	if r.InputValues == nil {
		r.InputValues = map[string]interface{}{}
	}
	return &r, nil
}
