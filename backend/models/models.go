package models

import (
	"database/sql"
	"fmt"
	"time"
)

type Conversation struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Message struct {
	ID             int       `json:"id"`
	ConversationID int       `json:"conversation_id"`
	Role           string    `json:"role"`
	Content        string    `json:"content"`
	CreatedAt      time.Time `json:"created_at"`
}

func CreateConversation(db *sql.DB, title string) (*Conversation, error) {
	var c Conversation
	err := db.QueryRow(
		`INSERT INTO conversations (title) VALUES ($1) RETURNING id, title, created_at, updated_at`,
		title,
	).Scan(&c.ID, &c.Title, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create conversation: %w", err)
	}
	return &c, nil
}

func ListConversations(db *sql.DB) ([]Conversation, error) {
	rows, err := db.Query(
		`SELECT id, title, created_at, updated_at FROM conversations ORDER BY updated_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("list conversations: %w", err)
	}
	defer rows.Close()

	var convs []Conversation
	for rows.Next() {
		var c Conversation
		if err := rows.Scan(&c.ID, &c.Title, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		convs = append(convs, c)
	}
	return convs, nil
}

func GetConversation(db *sql.DB, id int) (*Conversation, error) {
	var c Conversation
	err := db.QueryRow(
		`SELECT id, title, created_at, updated_at FROM conversations WHERE id = $1`, id,
	).Scan(&c.ID, &c.Title, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get conversation: %w", err)
	}
	return &c, nil
}

func DeleteConversation(db *sql.DB, id int) error {
	_, err := db.Exec(`DELETE FROM conversations WHERE id = $1`, id)
	return err
}

func UpdateConversationTitle(db *sql.DB, id int, title string) error {
	_, err := db.Exec(
		`UPDATE conversations SET title = $1, updated_at = NOW() WHERE id = $2`, title, id,
	)
	return err
}

func UpdateConversationTimestamp(db *sql.DB, id int) error {
	_, err := db.Exec(`UPDATE conversations SET updated_at = NOW() WHERE id = $1`, id)
	return err
}

func CreateMessage(db *sql.DB, conversationID int, role, content string) (*Message, error) {
	var m Message
	err := db.QueryRow(
		`INSERT INTO messages (conversation_id, role, content) VALUES ($1, $2, $3) RETURNING id, conversation_id, role, content, created_at`,
		conversationID, role, content,
	).Scan(&m.ID, &m.ConversationID, &m.Role, &m.Content, &m.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create message: %w", err)
	}
	return &m, nil
}

func ListMessages(db *sql.DB, conversationID int) ([]Message, error) {
	rows, err := db.Query(
		`SELECT id, conversation_id, role, content, created_at FROM messages WHERE conversation_id = $1 ORDER BY created_at ASC`,
		conversationID,
	)
	if err != nil {
		return nil, fmt.Errorf("list messages: %w", err)
	}
	defer rows.Close()

	var msgs []Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.Role, &m.Content, &m.CreatedAt); err != nil {
			return nil, err
		}
		msgs = append(msgs, m)
	}
	return msgs, nil
}
