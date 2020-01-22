package models

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
)

// Conversation represents a conversation with one or more users
type Conversation struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

const (
	conversationsTable string = "conversations"
)

// CreateConversation adds a row to the "conversations" table
func (db *DB) CreateConversation(conversation *Conversation) (int64, error) {
	var b strings.Builder
	fmt.Fprintf(&b, "INSERT INTO %s(Name, Description) ", conversationsTable)
	fmt.Fprintf(&b, "VALUES(%q, %q)", conversation.Name, *conversation.Description)

	res, err := db.Exec(b.String())
	if err != nil {
		return -1, err
	}
	return res.LastInsertId()
}

// GetConversation queries for a single row from the "conversations" table
func (db *DB) GetConversation(id int64) (*Conversation, error) {
	conversation := &Conversation{}
	queryString := fmt.Sprintf("SELECT * FROM %s WHERE ID = '%d'", conversationsTable, id)

	err := db.QueryRow(queryString).Scan(&(conversation.ID), &(conversation.Name), &(conversation.Description))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return conversation, nil
}

// UpdateConversation updates an existing row in the "conversations" table
func (db *DB) UpdateConversation(conversation *Conversation) error {
	if conversation.Name == "" && conversation.Description == nil {
		return nil
	}

	var b strings.Builder
	fmt.Fprintf(&b, "UPDATE %s SET ", conversationsTable)
	if conversation.Name != "" {
		fmt.Fprintf(&b, "Name=%q ", conversation.Name)
	}
	if conversation.Description != nil {
		fmt.Fprintf(&b, "Description=%q ", *conversation.Description)
	}
	fmt.Fprintf(&b, "WHERE ID=%d", conversation.ID)

	_, err := db.Exec(b.String())
	return err
}

// DeleteConversation removes a row from the "conversations" table
func (db *DB) DeleteConversation(id int64) error {
	queryString := fmt.Sprintf("DELETE FROM %s WHERE ID = '%d'", conversationsTable, id)

	res, err := db.Exec(queryString)
	if err == nil {
		if rowCount, err := res.RowsAffected(); err == nil {
			log.Printf("Deleted %d row(s) from \"%s\"", rowCount, conversationsTable)
		}
	}
	return err
}
