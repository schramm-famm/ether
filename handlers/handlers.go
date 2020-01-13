package handlers

import (
	"encoding/json"
	"net/http"
)

// PostConversationHandler creates a new conversation
func PostConversationHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	reqBody := models.NewConversation()
	if err := parseReqBody(w, r.Body, reqBody); err != nil {
		return
	}

	json.NewEncoder(w).Encode(reqBody)
}
