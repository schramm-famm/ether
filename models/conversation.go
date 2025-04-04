package models

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
)

// Conversation represents a conversation with one or more users
type Conversation struct {
	ID           int64   `json:"id"`
	Name         string  `json:"name"`
	Description  *string `json:"description"`
	AvatarURL    *string `json:"avatar_url"`
	LastModified string  `json:"last_modified"`
}

const (
	conversationsTable string = "conversations"
)

// Merge creates a new Conversation by copying the original Conversation and
// replacing its fields with the non-zero-value fields of a patch Conversation
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

	if patch.AvatarURL != nil {
		newConversation.AvatarURL = patch.AvatarURL
	} else {
		newConversation.AvatarURL = c.AvatarURL
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
	fmt.Fprintf(&b, "INSERT INTO %s(Name, Description, AvatarURL) ", conversationsTable)
	fmt.Fprintf(&b, "VALUES(?, ?, ?)")
	res, err := tx.Exec(b.String(), conversation.Name, *conversation.Description, *conversation.AvatarURL)
	if err != nil {
		tx.Rollback()
		return -1, err
	}

	if rowCount, err := res.RowsAffected(); err == nil {
		log.Printf(`Created %d row(s) in "%s"`, rowCount, conversationsTable)
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
	fmt.Fprintf(&b, "VALUES(?, ?, ?, ?, ?)")
	res, err = tx.Exec(b.String(), creatorID, conversationID, Owner, "", 0)
	if err != nil {
		tx.Rollback()
		return -1, err
	}

	if rowCount, err := res.RowsAffected(); err == nil {
		log.Printf(`Created %d row(s) in "%s"`, rowCount, mappingsTable)
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
	queryString := fmt.Sprintf("SELECT * FROM %s WHERE ID=?", conversationsTable)
	err := db.QueryRow(queryString, id).Scan(&(conversation.ID), &(conversation.Name), &(conversation.Description), &(conversation.AvatarURL), &(conversation.LastModified))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	log.Printf(`Read 1 row from "%s"`, conversationsTable)
	return conversation, nil
}

// GetConversations returns all conversations of a user
func (db *DB) GetConversations(userID int64, sort string) ([]Conversation, error) {
	var queryString strings.Builder
	fmt.Fprintf(&queryString, "SELECT c.ID, c.Name, c.Description, c.AvatarURL, c.LastModified ")
	fmt.Fprintf(&queryString, "FROM %s AS c JOIN %s AS m ON c.ID = m.ConversationID ", conversationsTable, mappingsTable)
	fmt.Fprintf(&queryString, "WHERE UserID=? ORDER BY c.LastModified ")
	if sort == "desc" || sort == "" {
		fmt.Fprintf(&queryString, "DESC")
	}
	rows, err := db.Query(queryString.String(), userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Create list of conversations
	c := Conversation{}
	conversations := make([]Conversation, 0)
	for rows.Next() {
		err := rows.Scan(&c.ID, &c.Name, &c.Description, &c.AvatarURL, &c.LastModified)
		if err != nil {
			return nil, err
		}
		conversations = append(conversations, c)
	}

	log.Printf("Read %d row(s)", len(conversations))
	return conversations, err
}

// UpdateConversation updates an existing row in the "conversations" table
func (db *DB) UpdateConversation(conversation *Conversation) error {
	if conversation.Name == "" && conversation.Description == nil && conversation.AvatarURL == nil {
		return nil
	}

	var b strings.Builder
	fmt.Fprintf(&b, "UPDATE %s SET ", conversationsTable)
	fmt.Fprintf(&b, "Name=?, Description=?, AvatarURL=? WHERE ID=?")
	res, err := db.Exec(b.String(), conversation.Name, *conversation.Description, *conversation.AvatarURL, conversation.ID)
	if err != nil {
		return err
	}

	if rowCount, err := res.RowsAffected(); err == nil {
		log.Printf(`Updated %d row(s) in "%s"`, rowCount, conversationsTable)
	} else {
		log.Println("Failed to get number of rows affected: " + err.Error())
	}
	return nil
}

// TouchConversation sets the LastModified value of a "conversations" table row
// to the current time.
func (db *DB) TouchConversation(conversationID int64) error {
	var b strings.Builder
	fmt.Fprintf(&b, "UPDATE %s SET ", conversationsTable)
	fmt.Fprintf(&b, "LastModified=NOW() WHERE ID=?")
	res, err := db.Exec(b.String(), conversationID)
	if err != nil {
		return err
	}

	if rowCount, err := res.RowsAffected(); err == nil {
		log.Printf(`Updated %d row(s) in "%s"`, rowCount, conversationsTable)
	} else {
		log.Println("Failed to get number of rows affected: " + err.Error())
	}
	return nil
}

// DeleteConversation removes a row from the "conversations" table
func (db *DB) DeleteConversation(id int64) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	queryString := fmt.Sprintf("DELETE FROM %s WHERE ConversationID=?", mappingsTable)
	res, err := tx.Exec(queryString, id)
	if err != nil {
		tx.Rollback()
		return err
	}

	if rowCount, err := res.RowsAffected(); err == nil {
		log.Printf(`Deleted %d row(s) from "%s"`, rowCount, mappingsTable)
	} else {
		tx.Rollback()
		return err
	}

	queryString = fmt.Sprintf("DELETE FROM %s WHERE ID=?", conversationsTable)
	res, err = tx.Exec(queryString, id)
	if err != nil {
		tx.Rollback()
		return err
	}

	if rowCount, err := res.RowsAffected(); err == nil {
		log.Printf(`Deleted %d row(s) from "%s"`, rowCount, conversationsTable)
	} else {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	return err
}
