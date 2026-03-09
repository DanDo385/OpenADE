package services

import (
	"context"
	"os/exec"
	"strings"
	"time"

	"openade/internal/model"
)

// Allowed command families
var allowedCommands = map[string]bool{
	"help":   true,
	"run":    true,
	"export": true,
	"import": true,
	"echo":   true,
	"date":   true,
	"pwd":    true,
}

type CommandService struct{}

func NewCommandService() *CommandService {
	return &CommandService{}
}

func (s *CommandService) Execute(ctx context.Context, req model.CommandExecuteRequest) model.CommandExecuteResponse {
	start := time.Now()

	if !req.Confirm {
		return model.CommandExecuteResponse{
			OK:         false,
			Output:     "",
			ExitCode:   1,
			DurationMs: time.Since(start).Milliseconds(),
		}
	}

	input := strings.TrimSpace(req.Input)
	if input == "" {
		return model.CommandExecuteResponse{
			OK:         false,
			Output:     "input is required",
			ExitCode:   1,
			DurationMs: time.Since(start).Milliseconds(),
		}
	}

	// Strip leading / if present
	if len(input) > 0 && input[0] == '/' {
		input = strings.TrimSpace(input[1:])
	}

	parts := strings.Fields(input)
	if len(parts) == 0 {
		return model.CommandExecuteResponse{
			OK:         false,
			Output:     "command is required",
			ExitCode:   1,
			DurationMs: time.Since(start).Milliseconds(),
		}
	}

	cmdName := strings.ToLower(parts[0])
	args := parts[1:]

	// Built-in commands (no exec)
	switch cmdName {
	case "help":
		return model.CommandExecuteResponse{
			OK: true,
			Output: `OpenADE Commands:
  /help      - Show this help
  /run       - Run a task (use Run panel)
  /export    - Export task
  /import    - Import task
  /agent:X   - Launch agent (e.g. /agent:blackjack)`,
			ExitCode:   0,
			DurationMs: time.Since(start).Milliseconds(),
		}
	case "run", "export", "import":
		return model.CommandExecuteResponse{
			OK:         true,
			Output:     "Use the Tasks panel for run, export, and import.",
			ExitCode:   0,
			DurationMs: time.Since(start).Milliseconds(),
		}
	}

	// Exec allowlist: only echo, date, pwd (no shell, no user input to exec)
	if !allowedCommands[cmdName] {
		return model.CommandExecuteResponse{
			OK:         false,
			Output:     "command not in allowlist: " + cmdName,
			ExitCode:   1,
			DurationMs: time.Since(start).Milliseconds(),
		}
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
		return model.CommandExecuteResponse{
			OK:         false,
			Output:     "command not executable: " + cmdName,
			ExitCode:   1,
			DurationMs: time.Since(start).Milliseconds(),
		}
	}

	out, err := cmd.CombinedOutput()
	dur := time.Since(start).Milliseconds()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = 1
		}
	}

	return model.CommandExecuteResponse{
		OK:         exitCode == 0,
		Output:     string(out),
		ExitCode:   exitCode,
		DurationMs: dur,
	}
}
