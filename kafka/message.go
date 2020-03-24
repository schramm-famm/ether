package kafka

// Message represents a json-encoded WebSocket message using the Riht protocol.
type Message struct {
	Type MessageType `json:"type"`
	Data InnerData   `json:"data"`
}

// MessageType represents the possible Riht WebSocket protocol message types.
type MessageType int

// UpdateType represents the possible Update message subtypes.
type UpdateType int

const (
	TypeInit      MessageType = 0
	TypeUpdate    MessageType = 1
	TypeAck       MessageType = 2
	TypeSync      MessageType = 3
	TypeUserJoin  MessageType = 4
	TypeUserLeave MessageType = 5

	UpdateTypeEdit   UpdateType = 0
	UpdateTypeCursor UpdateType = 1
)

// InnerData represents the payload of a Riht WebSocket protocol message.
type InnerData struct {
	Type        *UpdateType      `json:"type,omitempty"`
	Version     *int             `json:"version,omitempty"`
	Patch       *string          `json:"patch,omitempty"`
	Delta       *Delta           `json:"delta,omitempty"`
	UserID      *int64           `json:"user_id,omitempty"`
	Content     *string          `json:"content,omitempty"`
	ActiveUsers *map[int64]Caret `json:"active_users,omitempty"`
}

// Delta represents the caret and content shifts associated with an Update
// message.
type Delta struct {
	CaretStart *int `json:"caret_start,omitempty"`
	CaretEnd   *int `json:"caret_end,omitempty"`
	Doc        *int `json:"doc,omitempty"`
}

// Caret represents a user's Start and End position in the document
type Caret struct {
	Start int `json:"start"`
	End   int `json:"end"`
}
