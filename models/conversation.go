package models

// Conversation represents a conversation with one or more users
type Conversation struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	FilePath    string `json:"file_path,omitempty"`
	Picture     string `json:"picture,omitempty"`
}
