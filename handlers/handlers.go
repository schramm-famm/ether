package handlers

import (
	"encoding/json"
	"ether/models"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
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

// PostConversationsHandler creates a new conversation
func PostConversationsHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	reqBody := models.NewConversation()
	if err := parseReqBody(w, r.Body, reqBody); err != nil {
		return
	}

	json.NewEncoder(w).Encode(reqBody)
}

// GetConversationsHandler gets filtered conversations for a user
func GetConversationsHandler(w http.ResponseWriter, r *http.Request) {
	queryValues, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		errMsg := "Failed to parse query: " + err.Error()
		log.Println(errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	resBody := map[string]string{
		"user_id": queryValues.Get("user_id"),
	}

	json.NewEncoder(w).Encode(resBody)
}

// PutConversationsHandler replaces, or creates if does not exist, a conversation
func PutConversationsHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	reqBody := models.NewConversation()
	if err := parseReqBody(w, r.Body, reqBody); err != nil {
		return
	}

	json.NewEncoder(w).Encode(reqBody)
}

// DeleteConversationsHandler deletes a conversation
func DeleteConversationsHandler(w http.ResponseWriter, r *http.Request) {
	queryValues, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		errMsg := "Failed to parse query: " + err.Error()
		log.Println(errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	resBody := map[string]string{
		"user_id": queryValues.Get("user_id"),
	}

	json.NewEncoder(w).Encode(resBody)
}

// PatchConversationsHandler updates a conversation
func PatchConversationsHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	reqBody := models.NewConversation()
	if err := parseReqBody(w, r.Body, reqBody); err != nil {
		return
	}

	json.NewEncoder(w).Encode(reqBody)
}
