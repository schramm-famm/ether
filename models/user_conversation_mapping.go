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
func (db *DB) CreateUserConversationMapping(mapping *UserConversationMapping) error {
	var b strings.Builder
	fmt.Fprintf(&b, "INSERT INTO %s(UserID, ConversationID, Role, Nickname, Pending) ", mappingsTable)
	fmt.Fprintf(&b, "VALUES(?, ?, ?, ?, ?)")
	pendingFlag := 0
	if mapping.Pending {
		pendingFlag = 1
	}
	res, err := db.Exec(
		b.String(),
		mapping.UserID,
		mapping.ConversationID,
		mapping.Role,
		mapping.Nickname,
		pendingFlag,
	)
	if err != nil {
		return err
	}
	if rowCount, err := res.RowsAffected(); err == nil {
		log.Printf(`Created %d row(s) in "%s"`, rowCount, mappingsTable)
	}
	return nil
}

// GetUserConversationMapping queries for a single row from the
// "users_to_conversations" table using the combination of ConversationID and
// UserID which should be unique to each row
func (db *DB) GetUserConversationMapping(userID, conversationID int64) (*UserConversationMapping, error) {
	var tmpPending string
	mapping := &UserConversationMapping{}
	queryString := fmt.Sprintf("SELECT * FROM %s WHERE UserID=? AND ConversationID=?", mappingsTable)
	err := db.QueryRow(queryString, userID, conversationID).Scan(
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
	log.Printf(`Read 1 row from "%s"`, mappingsTable)
	return mapping, nil
}

// UpdateUserConversationMapping updates an existing row in the
// "users_to_conversations" table
func (db *DB) UpdateUserConversationMapping(mapping *UserConversationMapping) error {
	var b strings.Builder
	fmt.Fprintf(&b, "UPDATE %s SET ", mappingsTable)
	fmt.Fprintf(&b, "Role=?, Nickname=?, Pending=? ")
	fmt.Fprintf(&b, "WHERE UserID=? AND ConversationID=?")
	pendingFlag := 0
	if mapping.Pending {
		pendingFlag = 1
	}
	res, err := db.Exec(
		b.String(),
		mapping.Role,
		mapping.Nickname,
		pendingFlag,
		mapping.UserID,
		mapping.ConversationID,
	)
	if err != nil {
		return err
	}

	if rowCount, err := res.RowsAffected(); err == nil {
		log.Printf(`Updated %d row(s) in "%s"`, rowCount, mappingsTable)
	} else {
		log.Println("Failed to get number of rows affected: " + err.Error())
	}
	return nil
}

// DeleteUserConversationMapping removes a row from the "users_to_conversations"
// table
func (db *DB) DeleteUserConversationMapping(userID, conversationID int64) error {
	queryString := fmt.Sprintf("DELETE FROM %s WHERE UserID=? AND ConversationID=?", mappingsTable)
	res, err := db.Exec(queryString, userID, conversationID)
	if err != nil {
		return err
	}

	if rowCount, err := res.RowsAffected(); err == nil {
		log.Printf(`Deleted %d row(s) in "%s"`, rowCount, mappingsTable)
	} else {
		log.Println("Failed to get number of rows deleted: " + err.Error())
	}
	return nil
}
