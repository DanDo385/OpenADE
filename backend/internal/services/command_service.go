package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"openade/internal/llm"
	"openade/internal/model"
	"openade/internal/secrets"
)

var builtinCommands = map[string]bool{
	"help":        true,
	"run":         true,
	"export":      true,
	"import":      true,
	"summarize":   true,
	"list-runs":   true,
	"inspect-run": true,
	"objective":   true,
}

var allowedExecCommands = map[string]bool{
	"echo": true,
	"date": true,
	"pwd":  true,
}

type CommandService struct {
	Providers     *ProviderService
	Conversations *ConversationService
	Runs          *RunService
	Secrets       secrets.Provider
	NewAdapter    func(cfg *model.ProviderConfig) llm.Adapter
}

func NewCommandService(
	providers *ProviderService,
	conversations *ConversationService,
	runs *RunService,
	secretProvider secrets.Provider,
	newAdapter func(cfg *model.ProviderConfig) llm.Adapter,
) *CommandService {
	return &CommandService{
		Providers:     providers,
		Conversations: conversations,
		Runs:          runs,
		Secrets:       secretProvider,
		NewAdapter:    newAdapter,
	}
}

func (s *CommandService) Execute(ctx context.Context, req model.CommandExecuteRequest) model.CommandExecuteResponse {
	start := time.Now()

	if !req.Confirm {
		return model.CommandExecuteResponse{
			OK:         false,
			ExitCode:   1,
			DurationMs: time.Since(start).Milliseconds(),
			Output:     "confirmation is required",
		}
	}

	input := strings.TrimSpace(req.Input)
	if input == "" {
		return model.CommandExecuteResponse{
			OK:         false,
			ExitCode:   1,
			DurationMs: time.Since(start).Milliseconds(),
			Output:     "input is required",
		}
	}

	if strings.HasPrefix(input, "/") {
		input = strings.TrimSpace(input[1:])
	}

	parts := strings.Fields(input)
	if len(parts) == 0 {
		return model.CommandExecuteResponse{
			OK:         false,
			ExitCode:   1,
			DurationMs: time.Since(start).Milliseconds(),
			Output:     "command is required",
		}
	}

	cmdName := strings.ToLower(parts[0])
	args := parts[1:]

	if builtinCommands[cmdName] {
		resp := s.executeBuiltin(ctx, cmdName, args)
		resp.DurationMs = time.Since(start).Milliseconds()
		return resp
	}

	resp := s.executeShell(ctx, cmdName, args)
	resp.DurationMs = time.Since(start).Milliseconds()
	return resp
}

func (s *CommandService) executeBuiltin(ctx context.Context, cmdName string, args []string) model.CommandExecuteResponse {
	switch cmdName {
	case "help":
		return model.CommandExecuteResponse{
			OK:       true,
			ExitCode: 0,
			Output: `OpenADE Commands:
  /help                    - Show this help
  /run                     - Use the Tasks panel to run a task
  /export                  - Use the Tasks panel to export a task
  /import                  - Use the Tasks panel to import a task
  /summarize <conversation_id>
                           - Summarize a conversation with the configured LLM
  /list-runs [task_id]     - List recent runs, optionally filtered by task
  /inspect-run <run_id>    - Show details for a run
  /objective <conversation_id>
                           - Draft an objective from a conversation
  /echo hello              - Echo text
  /date                    - Show current date
  /pwd                     - Show working directory`,
		}
	case "run", "export", "import":
		return model.CommandExecuteResponse{
			OK:       true,
			ExitCode: 0,
			Output:   "Use the Tasks panel for run, export, and import.",
		}
	case "summarize":
		if len(args) < 1 {
			return failureResponse("usage: /summarize <conversation_id>")
		}
		output, err := s.summarizeConversation(ctx, args[0])
		if err != nil {
			return failureResponse(err.Error())
		}
		return successResponse(output)
	case "list-runs":
		taskID := ""
		if len(args) > 0 {
			taskID = args[0]
		}
		output, err := s.listRuns(ctx, taskID)
		if err != nil {
			return failureResponse(err.Error())
		}
		return successResponse(output)
	case "inspect-run":
		if len(args) < 1 {
			return failureResponse("usage: /inspect-run <run_id>")
		}
		output, err := s.inspectRun(ctx, args[0])
		if err != nil {
			return failureResponse(err.Error())
		}
		return successResponse(output)
	case "objective":
		if len(args) < 1 {
			return failureResponse("usage: /objective <conversation_id>")
		}
		output, err := s.draftObjective(ctx, args[0])
		if err != nil {
			return failureResponse(err.Error())
		}
		return successResponse(output)
	default:
		return failureResponse("unsupported built-in command: " + cmdName)
	}
}

func (s *CommandService) executeShell(ctx context.Context, cmdName string, args []string) model.CommandExecuteResponse {
	if !allowedExecCommands[cmdName] {
		return failureResponse("command not in allowlist: " + cmdName)
	}

	var cmd *exec.Cmd
	switch cmdName {
	case "echo":
		cmd = exec.CommandContext(ctx, "echo", args...)
	case "date":
		cmd = exec.CommandContext(ctx, "date", args...)
	case "pwd":
		cmd = exec.CommandContext(ctx, "pwd")
	default:
		return failureResponse("command not executable: " + cmdName)
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		exitCode := 1
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
		return model.CommandExecuteResponse{
			OK:       false,
			ExitCode: exitCode,
			Output:   string(out),
		}
	}

	return successResponse(string(out))
}

func (s *CommandService) summarizeConversation(ctx context.Context, conversationID string) (string, error) {
	messages, err := s.loadConversationMessages(ctx, conversationID)
	if err != nil {
		return "", err
	}

	adapter, modelName, err := s.getLLM(ctx)
	if err != nil {
		return "", err
	}

	llmMessages := []llm.ChatMessage{
		{
			Role:    "system",
			Content: "You summarize OpenADE conversations. Produce a concise, practical summary with the user's goal, important decisions, open questions, and next useful actions.",
		},
		{
			Role:    "user",
			Content: "Summarize this conversation:\n\n" + buildTranscript(messages),
		},
	}

	result, err := adapter.Complete(ctx, llmMessages, modelName)
	if err != nil {
		return "", fmt.Errorf("summarize llm call failed: %w", err)
	}
	return strings.TrimSpace(result.Content), nil
}

func (s *CommandService) listRuns(ctx context.Context, taskID string) (string, error) {
	if s.Runs == nil {
		return "", fmt.Errorf("run service is not configured")
	}
	runs, err := s.Runs.List(ctx, taskID)
	if err != nil {
		return "", err
	}
	if len(runs) == 0 {
		if taskID == "" {
			return "No runs found.", nil
		}
		return fmt.Sprintf("No runs found for task %s.", taskID), nil
	}

	var b strings.Builder
	for _, run := range runs {
		fmt.Fprintf(&b, "%s  task=%s  status=%s  model=%s  created=%s  duration=%dms  cost=$%.6f\n",
			run.ID, run.TaskID, run.Status, nonEmpty(run.Model, "n/a"), run.CreatedAt.Format(time.RFC3339), run.DurationMs, run.CostUSD)
	}
	return strings.TrimSpace(b.String()), nil
}

func (s *CommandService) inspectRun(ctx context.Context, runID string) (string, error) {
	if s.Runs == nil {
		return "", fmt.Errorf("run service is not configured")
	}
	run, err := s.Runs.Get(ctx, runID)
	if err != nil {
		return "", err
	}
	if run == nil {
		return "", fmt.Errorf("run not found")
	}

	inputsJSON, _ := json.MarshalIndent(run.InputValues, "", "  ")
	var b strings.Builder
	fmt.Fprintf(&b, "Run: %s\n", run.ID)
	fmt.Fprintf(&b, "Task ID: %s\n", run.TaskID)
	fmt.Fprintf(&b, "Task Version: %d\n", run.TaskVersion)
	fmt.Fprintf(&b, "Status: %s\n", run.Status)
	fmt.Fprintf(&b, "Model: %s\n", nonEmpty(run.Model, "n/a"))
	fmt.Fprintf(&b, "Created: %s\n", run.CreatedAt.Format(time.RFC3339))
	fmt.Fprintf(&b, "Duration: %dms\n", run.DurationMs)
	fmt.Fprintf(&b, "Cost USD: %.6f\n", run.CostUSD)
	fmt.Fprintf(&b, "Input Tokens: %d\n", run.InputTokens)
	fmt.Fprintf(&b, "Output Tokens: %d\n", run.OutputTokens)
	if run.Error != "" {
		fmt.Fprintf(&b, "Error: %s\n", run.Error)
	}
	fmt.Fprintf(&b, "\nInputs:\n%s\n", string(inputsJSON))
	fmt.Fprintf(&b, "\nPrompt:\n%s\n", run.PromptFinal)
	fmt.Fprintf(&b, "\nOutput:\n%s\n", run.Output)
	return strings.TrimSpace(b.String()), nil
}

func (s *CommandService) draftObjective(ctx context.Context, conversationID string) (string, error) {
	messages, err := s.loadConversationMessages(ctx, conversationID)
	if err != nil {
		return "", err
	}

	adapter, modelName, err := s.getLLM(ctx)
	if err != nil {
		return "", err
	}

	llmMessages := []llm.ChatMessage{
		{
			Role: "system",
			Content: `You draft OpenADE objectives from conversation transcripts.
Return only valid JSON with these keys:
- title
- goal
- constraints
- tools_required (array of strings)
- success_criteria`,
		},
		{
			Role:    "user",
			Content: "Draft an objective for this conversation:\n\n" + buildTranscript(messages),
		},
	}

	result, err := adapter.Complete(ctx, llmMessages, modelName)
	if err != nil {
		return "", fmt.Errorf("objective llm call failed: %w", err)
	}

	var draft model.UpsertObjectiveRequest
	content := stripCodeFences(result.Content)
	if err := json.Unmarshal([]byte(content), &draft); err != nil {
		return "", fmt.Errorf("failed to parse objective draft: %w", err)
	}

	return formatObjectiveDraft(draft), nil
}

func (s *CommandService) loadConversationMessages(ctx context.Context, conversationID string) ([]model.Message, error) {
	if s.Conversations == nil {
		return nil, fmt.Errorf("conversation service is not configured")
	}
	conversation, err := s.Conversations.Get(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	if conversation == nil {
		return nil, fmt.Errorf("conversation not found")
	}
	if len(conversation.Messages) == 0 {
		return nil, fmt.Errorf("conversation has no messages")
	}
	return conversation.Messages, nil
}

func (s *CommandService) getLLM(ctx context.Context) (llm.Adapter, string, error) {
	if s.Providers == nil || s.NewAdapter == nil {
		return nil, "", fmt.Errorf("command service is missing LLM dependencies")
	}
	cfg, err := s.Providers.GetDefault(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("loading provider config: %w", err)
	}
	if cfg == nil {
		return nil, "", fmt.Errorf("no LLM provider configured")
	}
	return s.NewAdapter(cfg), cfg.DefaultModel, nil
}

func buildTranscript(messages []model.Message) string {
	var b strings.Builder
	for _, msg := range messages {
		fmt.Fprintf(&b, "[%s] %s\n\n", msg.Role, msg.Content)
	}
	return strings.TrimSpace(b.String())
}

func formatObjectiveDraft(draft model.UpsertObjectiveRequest) string {
	var b strings.Builder
	title := nonEmpty(strings.TrimSpace(draft.Title), "Untitled Objective")
	fmt.Fprintf(&b, "# %s\n\n", title)
	if draft.Goal != "" {
		fmt.Fprintf(&b, "## Goal\n\n%s\n\n", strings.TrimSpace(draft.Goal))
	}
	if draft.Constraints != "" {
		fmt.Fprintf(&b, "## Constraints\n\n%s\n\n", strings.TrimSpace(draft.Constraints))
	}
	if len(draft.ToolsRequired) > 0 {
		b.WriteString("## Tools Required\n\n")
		for _, tool := range draft.ToolsRequired {
			fmt.Fprintf(&b, "- %s\n", tool)
		}
		b.WriteString("\n")
	}
	if draft.SuccessCriteria != "" {
		fmt.Fprintf(&b, "## Success Criteria\n\n%s\n", strings.TrimSpace(draft.SuccessCriteria))
	}
	return strings.TrimSpace(b.String())
}

func stripCodeFences(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "```json")
	value = strings.TrimPrefix(value, "```")
	value = strings.TrimSuffix(value, "```")
	return strings.TrimSpace(value)
}

func successResponse(output string) model.CommandExecuteResponse {
	return model.CommandExecuteResponse{
		OK:       true,
		ExitCode: 0,
		Output:   output,
	}
}

func failureResponse(output string) model.CommandExecuteResponse {
	return model.CommandExecuteResponse{
		OK:       false,
		ExitCode: 1,
		Output:   output,
	}
}

func nonEmpty(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
