package handlers

import (
	"encoding/json"
	"ether/models"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

func parseReqBody(w http.ResponseWriter, body io.ReadCloser, bodyObj *models.Conversation) error {
	bodyBytes, err := ioutil.ReadAll(body)
	if err != nil {
		errMsg := "Failed to read request body: " + err.Error()
		log.Println(errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return err
	}

	if err := json.Unmarshal(bodyBytes, bodyObj); err != nil {
		errMsg := "Failed to parse request body: " + err.Error()
		log.Println(errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return err
	}

	return nil
}

// PostConversationHandler creates a new conversation
func PostConversationHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	reqBody := models.NewConversation()
	if err := parseReqBody(w, r.Body, reqBody); err != nil {
		return
	}

	json.NewEncoder(w).Encode(reqBody)
}
