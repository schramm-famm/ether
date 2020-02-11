package models

type MockDB struct {
	Conversation *Conversation
	Mapping      *UserConversationMapping
	Mappings     []*UserConversationMapping
	Errors       []error

	Count           int
	ConversationIDs []int64
	UserIDs         []int64
}

func NewMockDB(
	conversation *Conversation,
	mapping *UserConversationMapping,
	mappings []*UserConversationMapping,
	errors []error,
) *MockDB {
	return &MockDB{
		Conversation:    conversation,
		Mapping:         mapping,
		Mappings:        mappings,
		Errors:          errors,
		Count:           0,
		ConversationIDs: make([]int64, 0),
		UserIDs:         make([]int64, 0),
	}
}

func (db *MockDB) getError() error {
	if db.Errors == nil {
		return nil
	} else {
		err := db.Errors[db.Count]
		db.Count++
		return err
	}
}

func (db *MockDB) CreateConversation(conversation *Conversation, creatorID int64) (int64, error) {
	db.Conversation = conversation
	db.Conversation.ID = 1
	db.UserIDs = append(db.UserIDs, creatorID)
	return db.Conversation.ID, db.getError()
}

func (db *MockDB) GetConversation(id int64) (*Conversation, error) {
	db.ConversationIDs = append(db.ConversationIDs, id)
	return db.Conversation, db.getError()
}

func (db *MockDB) UpdateConversation(conversation *Conversation) error {
	db.Conversation = conversation
	return db.getError()
}

func (db *MockDB) DeleteConversation(id int64) error {
	db.ConversationIDs = append(db.ConversationIDs, id)
	return db.getError()
}

func (db *MockDB) CreateUserConversationMapping(mapping *UserConversationMapping) error {
	db.Mapping = mapping
	return db.getError()
}

func (db *MockDB) GetUserConversationMapping(userID, conversationID int64) (*UserConversationMapping, error) {
	db.UserIDs = append(db.UserIDs, userID)
	db.ConversationIDs = append(db.ConversationIDs, conversationID)
	return db.Mapping, db.getError()
}

func (db *MockDB) GetUserConversationMappings(conversationID int64) ([]*UserConversationMapping, error) {
	db.ConversationIDs = append(db.ConversationIDs, conversationID)
	return db.Mappings, db.getError()
}

func (db *MockDB) UpdateUserConversationMapping(mapping *UserConversationMapping) error {
	db.Mapping = mapping
	return db.getError()
}

func (db *MockDB) DeleteUserConversationMapping(userID, conversationID int64) error {
	db.UserIDs = append(db.UserIDs, userID)
	db.ConversationIDs = append(db.ConversationIDs, conversationID)
	return db.getError()
}
