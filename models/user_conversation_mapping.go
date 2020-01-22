package models

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
)

type UserConversationMapping struct {
	UserID         int64  `json:"user_id,omitempty"`
	ConversationID int64  `json:"conversation_id,omitempty"`
	Role           Role   `json:"role,omitempty"`
	Nickname       string `json:"nickname,omitempty"`
	Pending        bool   `json:"pending,omitempty"`
	LastOpened     string `json:"last_opened,omitempty"`
}

type Role string

const (
	Owner Role = "owner"
	Admin Role = "admin"
	User  Role = "user"

	mappingsTable string = "users_to_conversations"
)

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
	return res.LastInsertId()
}

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
	return mapping, nil
}

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
