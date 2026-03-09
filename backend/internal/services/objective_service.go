package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"openade/internal/db"
	"openade/internal/model"
)

type ObjectiveService struct {
	DB *sql.DB
}

func NewObjectiveService(database *sql.DB) *ObjectiveService {
	return &ObjectiveService{DB: database}
}

func (s *ObjectiveService) GetByConversationID(ctx context.Context, conversationID string) (*model.Objective, error) {
	var obj model.Objective
	var createdAt, updatedAt, toolsJSON string

	err := s.DB.QueryRowContext(ctx,
		`SELECT id, conversation_id, title, goal, constraints, tools_required, success_criteria, created_at, updated_at
		 FROM objectives WHERE conversation_id = ?`,
		conversationID,
	).Scan(&obj.ID, &obj.ConversationID, &obj.Title, &obj.Goal, &obj.Constraints, &toolsJSON, &obj.SuccessCriteria, &createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting objective: %w", err)
	}

	obj.CreatedAt = db.ParseTime(createdAt)
	obj.UpdatedAt = db.ParseTime(updatedAt)

	if err := json.Unmarshal([]byte(toolsJSON), &obj.ToolsRequired); err != nil {
		obj.ToolsRequired = []string{}
	}

	return &obj, nil
}

func (s *ObjectiveService) Upsert(ctx context.Context, conversationID string, req model.UpsertObjectiveRequest) (*model.Objective, error) {
	id := uuid.NewString()
	now := db.FormatTime(time.Now())

	tools := req.ToolsRequired
	if tools == nil {
		tools = []string{}
	}
	toolsJSON, err := json.Marshal(tools)
	if err != nil {
		return nil, fmt.Errorf("marshaling tools_required: %w", err)
	}

	_, err = s.DB.ExecContext(ctx,
		`INSERT INTO objectives (id, conversation_id, title, goal, constraints, tools_required, success_criteria, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(conversation_id) DO UPDATE SET
			title = excluded.title,
			goal = excluded.goal,
			constraints = excluded.constraints,
			tools_required = excluded.tools_required,
			success_criteria = excluded.success_criteria,
			updated_at = excluded.updated_at`,
		id, conversationID, req.Title, req.Goal, req.Constraints, string(toolsJSON), req.SuccessCriteria, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("upserting objective: %w", err)
	}

	return s.GetByConversationID(ctx, conversationID)
}

func (s *ObjectiveService) ExportMarkdown(obj *model.Objective) string {
	var b strings.Builder

	fmt.Fprintf(&b, "# %s\n\n", obj.Title)

	if obj.Goal != "" {
		fmt.Fprintf(&b, "## Goal\n\n%s\n\n", obj.Goal)
	}

	if obj.Constraints != "" {
		fmt.Fprintf(&b, "## Constraints\n\n%s\n\n", obj.Constraints)
	}

	if len(obj.ToolsRequired) > 0 {
		b.WriteString("## Tools Required\n\n")
		for _, t := range obj.ToolsRequired {
			fmt.Fprintf(&b, "- %s\n", t)
		}
		b.WriteString("\n")
	}

	if obj.SuccessCriteria != "" {
		fmt.Fprintf(&b, "## Success Criteria\n\n%s\n", obj.SuccessCriteria)
	}

	return b.String()
}
