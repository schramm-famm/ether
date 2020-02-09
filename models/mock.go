package models

type MockDB struct {
	Conversation *Conversation
	Mapping      *UserConversationMapping
	Mappings     []*UserConversationMapping
	Error        error

	ConversationIDs []int64
	UserIDs         []int64
}

func (db *MockDB) CreateConversation(conversation *Conversation, creatorID int64) (int64, error) {
	db.Conversation = conversation
	db.Conversation.ID = 1
	db.UserIDs = append(db.UserIDs, creatorID)
	return db.Conversation.ID, db.Error
}

func (db *MockDB) GetConversation(id int64) (*Conversation, error) {
	db.ConversationIDs = append(db.ConversationIDs, id)
	return db.Conversation, db.Error
}

func (db *MockDB) UpdateConversation(conversation *Conversation) error {
	db.Conversation = conversation
	return db.Error
}

func (db *MockDB) DeleteConversation(id int64) error {
	db.ConversationIDs = append(db.ConversationIDs, id)
	return db.Error
}

func (db *MockDB) CreateUserConversationMapping(mapping *UserConversationMapping) error {
	db.Mapping = mapping
	return db.Error
}

func (db *MockDB) GetUserConversationMapping(userID, conversationID int64) (*UserConversationMapping, error) {
	db.UserIDs = append(db.UserIDs, userID)
	db.ConversationIDs = append(db.ConversationIDs, conversationID)
	return db.Mapping, db.Error
}

func (db *MockDB) GetUserConversationMappings(conversationID int64) ([]*UserConversationMapping, error) {
	db.ConversationIDs = append(db.ConversationIDs, conversationID)
	return db.Mappings, db.Error
}

func (db *MockDB) UpdateUserConversationMapping(mapping *UserConversationMapping) error {
	db.Mapping = mapping
	return db.Error
}

func (db *MockDB) DeleteUserConversationMapping(userID, conversationID int64) error {
	db.UserIDs = append(db.UserIDs, userID)
	db.ConversationIDs = append(db.ConversationIDs, conversationID)
	return db.Error
}
