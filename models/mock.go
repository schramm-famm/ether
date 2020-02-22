package models

import (
	"sort"
)

type MockDB struct {
	Conversations   map[int64]*Conversation
	Mappings        map[int64]map[int64]*UserConversationMapping
	Errors          []error
	Count           int
	AutoIncrementID int64
}

func NewMockDB(
	conversations []*Conversation,
	mappings []*UserConversationMapping,
	errors []error,
) *MockDB {
	db := &MockDB{
		Conversations:   make(map[int64]*Conversation),
		Mappings:        make(map[int64]map[int64]*UserConversationMapping),
		Errors:          errors,
		Count:           0,
		AutoIncrementID: 0,
	}
	for _, conversation := range conversations {
		if conversation == nil {
			continue
		}
		db.Conversations[conversation.ID] = conversation
		if conversation.ID > db.AutoIncrementID {
			db.AutoIncrementID = conversation.ID
		}
	}
	for _, mapping := range mappings {
		if mapping == nil {
			continue
		}
		db.SetMapping(mapping.UserID, mapping.ConversationID, mapping)
	}
	return db
}

func (db *MockDB) getError() error {
	if db.Errors == nil || db.Count >= len(db.Errors) {
		return nil
	}
	err := db.Errors[db.Count]
	db.Count++
	return err
}

func (db *MockDB) GetMapping(userID, conversationID int64) *UserConversationMapping {
	if db.Mappings[conversationID] == nil {
		return nil
	}
	return db.Mappings[conversationID][userID]
}

func (db *MockDB) SetMapping(userID, conversationID int64, mapping *UserConversationMapping) {
	if db.Mappings[conversationID] == nil {
		db.Mappings[conversationID] = make(map[int64]*UserConversationMapping)
	}
	db.Mappings[conversationID][userID] = mapping
}

func (db *MockDB) CreateConversation(conversation *Conversation, creatorID int64) (int64, error) {
	if err := db.getError(); err != nil {
		return -1, err
	}
	db.AutoIncrementID++
	conversation.ID = db.AutoIncrementID
	db.Conversations[conversation.ID] = conversation
	db.SetMapping(creatorID, conversation.ID, &UserConversationMapping{})
	return conversation.ID, nil
}

func (db *MockDB) GetConversation(id int64) (*Conversation, error) {
	if err := db.getError(); err != nil {
		return nil, err
	}
	return db.Conversations[id], nil
}

func (db *MockDB) UpdateConversation(conversation *Conversation) error {
	if err := db.getError(); err != nil {
		return err
	}
	db.Conversations[conversation.ID] = conversation
	return nil
}

func (db *MockDB) DeleteConversation(id int64) error {
	if err := db.getError(); err != nil {
		return err
	}
	db.Conversations[id] = nil
	return nil
}

func (db *MockDB) CreateUserConversationMapping(mapping *UserConversationMapping) error {
	if err := db.getError(); err != nil {
		return err
	}
	db.SetMapping(mapping.UserID, mapping.ConversationID, mapping)
	return nil
}

func (db *MockDB) GetUserConversationMapping(userID, conversationID int64) (*UserConversationMapping, error) {
	if err := db.getError(); err != nil {
		return nil, err
	}
	return db.GetMapping(userID, conversationID), nil
}

func (db *MockDB) GetUserConversationMappings(conversationID int64) ([]*UserConversationMapping, error) {
	if err := db.getError(); err != nil {
		return nil, err
	}
	mappings := db.Mappings[conversationID]
	res := make([]*UserConversationMapping, 0)
	for _, mapping := range mappings {
		res = append(res, mapping)
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].UserID < res[j].UserID
	})
	return res, nil
}

func (db *MockDB) UpdateUserConversationMapping(mapping *UserConversationMapping) error {
	if err := db.getError(); err != nil {
		return err
	}
	db.SetMapping(mapping.UserID, mapping.ConversationID, mapping)
	return nil
}

func (db *MockDB) DeleteUserConversationMapping(userID, conversationID int64) error {
	if err := db.getError(); err != nil {
		return err
	}
	db.SetMapping(userID, conversationID, nil)
	return nil
}
