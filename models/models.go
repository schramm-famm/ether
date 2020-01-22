package models

import (
	"database/sql"

	// MySQL database driver
	_ "github.com/go-sql-driver/mysql"
)

// Datastore defines the CRUD operations of models in the database
type Datastore interface {
	CreateConversation(conversation *Conversation) (int64, error)
	GetConversation(id int64) (*Conversation, error)
	DeleteConversation(id int64) error
	UpdateConversation(conversation *Conversation) error

	CreateUserConversationMapping(mapping *UserConversationMapping) (int64, error)
	GetUserConversationMapping(userID, conversationID int64) (*UserConversationMapping, error)
	DeleteConversationMappings(conversationID int64) error
}

// DB represents an SQL database connection
type DB struct {
	*sql.DB
}

// NewDB initializes a new DB
func NewDB(dataSourceName string) (*DB, error) {
	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return &DB{db}, nil
}
