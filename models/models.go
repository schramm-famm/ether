package models

type Conversation struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description"`
	FilePath    string `json:"file_path,omitempty"`
	Picture     string `json:"picture,omitempty"`
}

func NewConversation() *Conversation {
	return &Conversation{
		ID:          "",
		Name:        "",
		Description: "",
		FilePath:    "",
		Picture:     "",
	}
}
