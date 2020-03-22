package models

import (
	"database/sql"
	"io/ioutil"
	"strings"

	// MySQL database driver
	_ "github.com/go-sql-driver/mysql"
)

// Datastore defines the CRUD operations of models in the database
type Datastore interface {
	CreateConversation(conversation *Conversation, creatorID int64) (int64, error)
	GetConversation(id int64) (*Conversation, error)
	GetConversations(userID int64, sort string) ([]Conversation, error)
	UpdateConversation(conversation *Conversation) error
	TouchConversation(conversationID int64) error
	DeleteConversation(id int64) error

	CreateUserConversationMapping(mapping *UserConversationMapping) error
	GetUserConversationMapping(userID, conversationID int64) (*UserConversationMapping, error)
	GetUserConversationMappings(conversationID int64) ([]*UserConversationMapping, error)
	UpdateUserConversationMapping(mapping *UserConversationMapping) error
	DeleteUserConversationMapping(userID, conversationID int64) error
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
	if err = setupDB(db); err != nil {
		return nil, err
	}
	return &DB{db}, nil
}

// setupDB creates the necessary "ether" database and tables if they don't
// already exist and makes the db connection use the "ether" database.
func setupDB(db *sql.DB) error {
	sqlScript, err := ioutil.ReadFile("dbSchema.sql")
	if err != nil {
		return err
	}

	statements := strings.Split(string(sqlScript), ";")
	if len(statements) > 0 {
		statements = statements[:len(statements)-1]
	}

	for _, statement := range statements {
		_, err = db.Exec(statement)
		if err != nil {
			return err
		}
	}
	return nil
}
