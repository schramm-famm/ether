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

func (c *Conversation) Merge(patch *Conversation) *Conversation {
	newConversation := &Conversation{ID: c.ID}

	if patch.Name != "" {
		newConversation.Name = patch.Name
	} else {
		newConversation.Name = c.Name
	}

	if patch.Description != nil {
		newConversation.Description = patch.Description
	} else {
		newConversation.Description = c.Description
	}

	return newConversation
}

// CreateConversation adds a row to the "conversations" table
func (db *DB) CreateConversation(conversation *Conversation, creatorID int64) (int64, error) {
	tx, err := db.Begin()
	if err != nil {
		return -1, err
	}

	var b strings.Builder
	fmt.Fprintf(&b, "INSERT INTO %s(Name, Description) ", conversationsTable)
	fmt.Fprintf(&b, "VALUES(%q, %q)", conversation.Name, *conversation.Description)

	res, err := tx.Exec(b.String())
	if err != nil {
		tx.Rollback()
		return -1, err
	}
	if rowCount, err := res.RowsAffected(); err == nil {
		log.Printf("Created %d row(s) in \"%s\"", rowCount, conversationsTable)
	} else {
		tx.Rollback()
		return -1, err
	}

	conversationID, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return -1, err
	}

	b.Reset()
	fmt.Fprintf(&b, "INSERT INTO %s(UserID, ConversationID, Role, Nickname, Pending) ", mappingsTable)
	fmt.Fprintf(
		&b,
		"VALUES(%d, %d, %q, %q, %b)",
		creatorID,
		conversationID,
		Owner,
		"",
		0,
	)

	res, err = tx.Exec(b.String())
	if err != nil {
		tx.Rollback()
		return -1, err
	}
	if rowCount, err := res.RowsAffected(); err == nil {
		log.Printf("Created %d row(s) in \"%s\"", rowCount, mappingsTable)
	} else {
		tx.Rollback()
		return -1, err
	}

	err = tx.Commit()
	if err != nil {
		return -1, err
	}
	return conversationID, nil
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
	log.Printf("Read 1 row from \"%s\"", conversationsTable)
	return conversation, nil
}

// UpdateConversation updates an existing row in the "conversations" table
func (db *DB) UpdateConversation(conversation *Conversation) error {
	if conversation.Name == "" && conversation.Description == nil {
		return nil
	}

	var b strings.Builder
	fmt.Fprintf(&b, "UPDATE %s SET ", conversationsTable)
	fmt.Fprintf(&b, "Name=%q, ", conversation.Name)
	fmt.Fprintf(&b, "Description=%q ", *conversation.Description)
	fmt.Fprintf(&b, "WHERE ID=%d", conversation.ID)

	res, err := db.Exec(b.String())
	if err == nil {
		if rowCount, err := res.RowsAffected(); err == nil {
			log.Printf("Updated %d row(s) in \"%s\"", rowCount, conversationsTable)
		}
	}
	return err
}

// DeleteConversation removes a row from the "conversations" table
func (db *DB) DeleteConversation(id int64) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	queryString := fmt.Sprintf("DELETE FROM %s WHERE ConversationID = '%d'", mappingsTable, id)

	res, err := tx.Exec(queryString)
	if err == nil {
		if rowCount, err := res.RowsAffected(); err == nil {
			log.Printf("Deleted %d row(s) from \"%s\"", rowCount, mappingsTable)
		}
	} else {
		tx.Rollback()
		return err
	}

	queryString = fmt.Sprintf("DELETE FROM %s WHERE ID = '%d'", conversationsTable, id)

	res, err = tx.Exec(queryString)
	if err == nil {
		if rowCount, err := res.RowsAffected(); err == nil {
			log.Printf("Deleted %d row(s) from \"%s\"", rowCount, conversationsTable)
		}
	} else {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	return err
}
