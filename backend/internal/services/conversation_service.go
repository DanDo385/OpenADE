package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"openade/internal/db"
	"openade/internal/model"
)

type ConversationService struct {
	DB *sql.DB
}

func NewConversationService(database *sql.DB) *ConversationService {
	return &ConversationService{DB: database}
}

func (s *ConversationService) Create(ctx context.Context) (*model.Conversation, error) {
	id := uuid.NewString()
	now := db.FormatTime(time.Now())

	_, err := s.DB.ExecContext(ctx,
		`INSERT INTO conversations (id, title, created_at, updated_at) VALUES (?, '', ?, ?)`,
		id, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("creating conversation: %w", err)
	}

	return &model.Conversation{
		ID:        id,
		CreatedAt: db.ParseTime(now),
		UpdatedAt: db.ParseTime(now),
	}, nil
}

func (s *ConversationService) List(ctx context.Context) ([]model.Conversation, error) {
	rows, err := s.DB.QueryContext(ctx,
		`SELECT id, title, created_at, updated_at FROM conversations ORDER BY updated_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("listing conversations: %w", err)
	}
	defer rows.Close()

	var convs []model.Conversation
	for rows.Next() {
		var c model.Conversation
		var createdAt, updatedAt string
		if err := rows.Scan(&c.ID, &c.Title, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("scanning conversation: %w", err)
		}
		c.CreatedAt = db.ParseTime(createdAt)
		c.UpdatedAt = db.ParseTime(updatedAt)
		convs = append(convs, c)
	}
	return convs, rows.Err()
}

func (s *ConversationService) Get(ctx context.Context, id string) (*model.Conversation, error) {
	var c model.Conversation
	var createdAt, updatedAt string
	err := s.DB.QueryRowContext(ctx,
		`SELECT id, title, created_at, updated_at FROM conversations WHERE id = ?`, id,
	).Scan(&c.ID, &c.Title, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting conversation: %w", err)
	}
	c.CreatedAt = db.ParseTime(createdAt)
	c.UpdatedAt = db.ParseTime(updatedAt)

	messages, err := s.GetMessages(ctx, id)
	if err != nil {
		return nil, err
	}
	c.Messages = messages

	return &c, nil
}

func (s *ConversationService) Delete(ctx context.Context, id string) error {
	res, err := s.DB.ExecContext(ctx, `DELETE FROM conversations WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("deleting conversation: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("conversation not found")
	}
	return nil
}

func (s *ConversationService) GetMessages(ctx context.Context, conversationID string) ([]model.Message, error) {
	rows, err := s.DB.QueryContext(ctx,
		`SELECT id, conversation_id, role, content, created_at
		 FROM messages WHERE conversation_id = ? ORDER BY created_at ASC`,
		conversationID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing messages: %w", err)
	}
	defer rows.Close()

	var msgs []model.Message
	for rows.Next() {
		var m model.Message
		var createdAt string
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.Role, &m.Content, &createdAt); err != nil {
			return nil, fmt.Errorf("scanning message: %w", err)
		}
		m.CreatedAt = db.ParseTime(createdAt)
		msgs = append(msgs, m)
	}
	return msgs, rows.Err()
}

func (s *ConversationService) AddMessage(ctx context.Context, conversationID, role, content string) (*model.Message, error) {
	id := uuid.NewString()
	now := db.FormatTime(time.Now())

	_, err := s.DB.ExecContext(ctx,
		`INSERT INTO messages (id, conversation_id, role, content, created_at) VALUES (?, ?, ?, ?, ?)`,
		id, conversationID, role, content, now,
	)
	if err != nil {
		return nil, fmt.Errorf("adding message: %w", err)
	}

	// Update conversation timestamp and title (title only on first message)
	s.DB.ExecContext(ctx, `UPDATE conversations SET updated_at = ? WHERE id = ?`, now, conversationID)
	if role == "user" {
		title := content
		if len(title) > 100 {
			title = title[:100]
		}
		s.DB.ExecContext(ctx,
			`UPDATE conversations SET title = ? WHERE id = ? AND title = ''`,
			title, conversationID,
		)
	}

	return &model.Message{
		ID:             id,
		ConversationID: conversationID,
		Role:           role,
		Content:        content,
		CreatedAt:      db.ParseTime(now),
	}, nil
}
