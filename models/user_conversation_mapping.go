package models

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
)

// UserConversationMapping represents a user's relationship to a conversation
type UserConversationMapping struct {
	UserID         int64  `json:"user_id,omitempty"`
	ConversationID int64  `json:"conversation_id,omitempty"`
	Role           Role   `json:"role,omitempty"`
	Nickname       string `json:"nickname,omitempty"`
	Pending        bool   `json:"pending,omitempty"`
	LastOpened     string `json:"last_opened,omitempty"`
}

// Role represents a user's access control rights in a conversation
type Role string

const (
	// Owner is a role that only the original creator of a conversation can have
	// and represents the highest level of privilege
	Owner Role = "owner"

	// Admin is a role that multiple non-creator users in a conversation can
	// have and represents elevated privilege over regular users
	Admin Role = "admin"

	// User is a role that multiple non-creator users in a converation can have
	// and represents the lowest level of privilege
	User Role = "user"

	mappingsTable string = "users_to_conversations"
)

// CreateUserConversationMapping adds a row to the "users_to_conversations" table
func (db *DB) CreateUserConversationMapping(mapping *UserConversationMapping) (int64, error) {
	var b strings.Builder
	fmt.Fprintf(&b, "INSERT INTO %s(UserID, ConversationID, Role, Nickname, Pending) ", mappingsTable)
	pendingFlag := 0
	if mapping.Pending {
		pendingFlag = 1
	}
	fmt.Fprintf(
		&b,
		"VALUES(%d, %d, %q, %q, %b)",
		mapping.UserID,
		mapping.ConversationID,
		mapping.Role,
		mapping.Nickname,
		pendingFlag,
	)

	res, err := db.Exec(b.String())
	if err != nil {
		return -1, err
	}
	if rowCount, err := res.RowsAffected(); err == nil {
		log.Printf("Created %d row(s) in \"%s\"", rowCount, mappingsTable)
	}
	return res.LastInsertId()
}

// GetUserConversationMapping queries for a single row from the
// "users_to_conversations" table using the combination of ConversationID and
// UserID which should be unique to each row
func (db *DB) GetUserConversationMapping(userID, conversationID int64) (*UserConversationMapping, error) {
	var tmpPending string
	mapping := &UserConversationMapping{}
	queryString := fmt.Sprintf(
		"SELECT * FROM %s WHERE UserID = %d AND ConversationID = %d",
		mappingsTable,
		userID,
		conversationID,
	)

	err := db.QueryRow(queryString).Scan(
		&(mapping.UserID),
		&(mapping.ConversationID),
		&(mapping.Role),
		&(mapping.Nickname),
		&tmpPending,
		&(mapping.LastOpened),
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	mapping.Pending = tmpPending == "\x00"
	log.Printf("Read 1 row from \"%s\"", mappingsTable)
	return mapping, nil
}

// DeleteConversationMappings removes every row in the "users_to_conversations"
// table with a specified ConversationID
func (db *DB) DeleteConversationMappings(conversationID int64) error {
	queryString := fmt.Sprintf("DELETE FROM %s WHERE ConversationID = '%d'", mappingsTable, conversationID)

	res, err := db.Exec(queryString)
	if err == nil {
		if rowCount, err := res.RowsAffected(); err == nil {
			log.Printf("Deleted %d row(s) from \"%s\"", rowCount, mappingsTable)
		}
	}
	return err
}
