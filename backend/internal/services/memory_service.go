package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"openade/internal/db"
	"openade/internal/model"
)

type MemoryService struct {
	DB *sql.DB
}

func NewMemoryService(database *sql.DB) *MemoryService {
	return &MemoryService{DB: database}
}

// GetAll returns all memory entries for a task.
func (s *MemoryService) GetAll(ctx context.Context, taskID string) ([]model.MemoryEntry, error) {
	rows, err := s.DB.QueryContext(ctx,
		`SELECT task_id, key, value, updated_at FROM memory WHERE task_id = ? ORDER BY key`,
		taskID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing memory: %w", err)
	}
	defer rows.Close()

	var entries []model.MemoryEntry
	for rows.Next() {
		var e model.MemoryEntry
		var updatedAt string
		if err := rows.Scan(&e.TaskID, &e.Key, &e.Value, &updatedAt); err != nil {
			return nil, fmt.Errorf("scanning memory entry: %w", err)
		}
		e.UpdatedAt = db.ParseTime(updatedAt)
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// GetMap returns all memory entries for a task as a simple map.
func (s *MemoryService) GetMap(ctx context.Context, taskID string) (map[string]string, error) {
	entries, err := s.GetAll(ctx, taskID)
	if err != nil {
		return nil, err
	}
	m := make(map[string]string)
	for _, e := range entries {
		m[e.Key] = e.Value
	}
	return m, nil
}

// Set upserts a single memory key-value for a task.
func (s *MemoryService) Set(ctx context.Context, taskID, key, value string) error {
	now := db.FormatTime(time.Now())
	_, err := s.DB.ExecContext(ctx,
		`INSERT INTO memory (task_id, key, value, updated_at) VALUES (?, ?, ?, ?)
		 ON CONFLICT(task_id, key) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at`,
		taskID, key, value, now,
	)
	if err != nil {
		return fmt.Errorf("setting memory key: %w", err)
	}
	return nil
}

// SetAll replaces all memory entries for a task with the provided map.
func (s *MemoryService) SetAll(ctx context.Context, taskID string, entries map[string]string) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete existing entries
	if _, err := tx.ExecContext(ctx, `DELETE FROM memory WHERE task_id = ?`, taskID); err != nil {
		return fmt.Errorf("clearing memory: %w", err)
	}

	now := db.FormatTime(time.Now())
	for key, value := range entries {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO memory (task_id, key, value, updated_at) VALUES (?, ?, ?, ?)`,
			taskID, key, value, now,
		); err != nil {
			return fmt.Errorf("inserting memory entry: %w", err)
		}
	}

	return tx.Commit()
}

// Delete removes a single memory key for a task.
func (s *MemoryService) Delete(ctx context.Context, taskID, key string) error {
	_, err := s.DB.ExecContext(ctx, `DELETE FROM memory WHERE task_id = ? AND key = ?`, taskID, key)
	if err != nil {
		return fmt.Errorf("deleting memory key: %w", err)
	}
	return nil
}
