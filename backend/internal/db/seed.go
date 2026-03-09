package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// SeedAgents inserts default game agents if the agents table is empty.
func SeedAgents(ctx context.Context, database *sql.DB) error {
	var count int
	if err := database.QueryRowContext(ctx, `SELECT COUNT(*) FROM agents`).Scan(&count); err != nil || count > 0 {
		return nil
	}

	now := FormatTime(time.Now())
	agents := []struct {
		name        string
		slug        string
		desc        string
		instructions string
	}{
		{"Blackjack", "blackjack", "A card game agent", "Play blackjack. Follow standard rules."},
		{"Trivia", "trivia", "A trivia quiz agent", "Answer trivia questions. Keep score."},
	}

	for _, a := range agents {
		id := uuid.NewString()
		bundle, _ := json.Marshal(map[string]any{})
		_, err := database.ExecContext(ctx,
			`INSERT INTO agents (id, name, slug, description, instructions, script_bundle_json, enabled, created_at, updated_at)
			 VALUES (?, ?, ?, ?, ?, ?, 1, ?, ?)`,
			id, a.name, a.slug, a.desc, a.instructions, string(bundle), now, now,
		)
		if err != nil {
			return err
		}
	}
	return nil
}
