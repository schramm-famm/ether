package handlers

import (
	"encoding/json"
	"ether/models"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// Env represents all application-level items that are needed by handlers
type Env struct {
	DB models.Datastore
}

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
func (env *Env) PostConversationsHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	userID, err := strconv.ParseInt(r.Header.Get("User-ID"), 10, 64)
	if err != nil {
		errMsg := "Invalid user ID"
		log.Println(errMsg + ": " + err.Error())
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	reqConversation := &models.Conversation{}
	if err := parseReqBody(w, r.Body, reqConversation); err != nil {
		return
	}

	conversationID, err := env.DB.CreateConversation(reqConversation)
	if err != nil {
		errMsg := "Failed to create row in \"conversations\" table" + err.Error()
		log.Println(errMsg)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	owner := &models.UserConversationMapping{
		UserID:         userID,
		ConversationID: conversationID,
		Role:           models.Owner,
		Pending:        false,
	}

	_, err = env.DB.CreateUserConversationMapping(owner)
	if err != nil {
		errMsg := "Failed to create row in \"user_to_conversations\" table" + err.Error()
		log.Println(errMsg)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// TODO: Create HTML file

	reqConversation.ID = conversationID
	json.NewEncoder(w).Encode(reqConversation)
}

// GetConversationsHandler gets filtered conversations for a user
func (env *Env) GetConversationsHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(r.Header.Get("User-ID"), 10, 64)
	if err != nil {
		errMsg := "Invalid user ID"
		log.Println(errMsg + ": " + err.Error())
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	conversationID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		errMsg := "Invalid ID"
		log.Println(errMsg + ": " + err.Error())
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	conversation, err := env.DB.GetConversation(conversationID)
	if err != nil {
		errMsg := "Internal Server Error"
		log.Println(errMsg + ": " + err.Error())
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	if conversation == nil {
		errMsg := fmt.Sprintf("Conversation %d does not exist", conversationID)
		log.Println(errMsg)
		http.Error(w, "Conversation not found", http.StatusNotFound)
		return
	}

	mapping, err := env.DB.GetUserConversationMapping(userID, conversationID)
	if err != nil {
		errMsg := "Internal Server Error"
		log.Println(errMsg + ": " + err.Error())
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	if mapping == nil {
		errMsg := fmt.Sprintf("User %d is not in conversation %d", userID, conversationID)
		log.Println(errMsg)
		http.Error(w, "Conversation not found", http.StatusNotFound)
		return
	}

	// TODO: Get conversation HTML
	// TODO: Get conversation picture

	json.NewEncoder(w).Encode(conversation)
}

// DeleteConversationsHandler deletes a conversation
func (env *Env) DeleteConversationsHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(r.Header.Get("User-ID"), 10, 64)
	if err != nil {
		errMsg := "Invalid user ID"
		log.Println(errMsg + ": " + err.Error())
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	conversationID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		errMsg := "Invalid ID"
		log.Println(errMsg + ": " + err.Error())
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	conversation, err := env.DB.GetConversation(conversationID)
	if err != nil {
		errMsg := "Internal Server Error"
		log.Println(errMsg + ": " + err.Error())
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	if conversation == nil {
		errMsg := fmt.Sprintf("Conversation %d does not exist", conversationID)
		log.Println(errMsg)
		http.Error(w, "Conversation not found", http.StatusNotFound)
		return
	}

	mapping, err := env.DB.GetUserConversationMapping(userID, conversationID)
	if err != nil {
		errMsg := "Internal Server Error"
		log.Println(errMsg + ": " + err.Error())
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	if mapping == nil {
		errMsg := fmt.Sprintf("User %d is not in conversation %d", userID, conversationID)
		log.Println(errMsg)
		http.Error(w, "Conversation not found", http.StatusNotFound)
		return
	}

	if mapping.Role != models.Owner {
		errMsg := fmt.Sprintf("User %d is not an Owner of conversation %d and cannot delete it", userID, conversationID)
		log.Println(errMsg)
		http.Error(w, "Forbidden from deleting conversation", http.StatusForbidden)
		return
	}

	err = env.DB.DeleteConversationMappings(conversationID)
	if err != nil {
		errMsg := "Internal Server Error"
		log.Println(errMsg + ": " + err.Error())
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	err = env.DB.DeleteConversation(conversationID)
	if err != nil {
		errMsg := "Internal Server Error"
		log.Println(errMsg + ": " + err.Error())
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(204)
}

// PatchConversationsHandler updates a conversation
func (env *Env) PatchConversationsHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	userID, err := strconv.ParseInt(r.Header.Get("User-ID"), 10, 64)
	if err != nil {
		errMsg := "Invalid user ID"
		log.Println(errMsg + ": " + err.Error())
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	reqBody := &models.Conversation{}
	if err := parseReqBody(w, r.Body, reqBody); err != nil {
		return
	}

	if reqBody.Name == "" && reqBody.Description == nil {
		errMsg := "Body has neither \"name\" nor \"description\""
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	conversationID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		errMsg := "Invalid ID"
		log.Println(errMsg + ": " + err.Error())
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	conversation, err := env.DB.GetConversation(conversationID)
	if err != nil {
		errMsg := "Internal Server Error"
		log.Println(errMsg + ": " + err.Error())
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	if conversation == nil {
		errMsg := fmt.Sprintf("Conversation %d does not exist", conversationID)
		log.Println(errMsg)
		http.Error(w, "Conversation not found", http.StatusNotFound)
		return
	}

	mapping, err := env.DB.GetUserConversationMapping(userID, conversationID)
	if err != nil {
		errMsg := "Internal Server Error"
		log.Println(errMsg + ": " + err.Error())
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	if mapping == nil {
		errMsg := fmt.Sprintf("User %d is not in conversation %d", userID, conversationID)
		log.Println(errMsg)
		http.Error(w, "Conversation not found", http.StatusNotFound)
		return
	}

	reqBody.ID = conversationID
	err = env.DB.UpdateConversation(reqBody)
	if err != nil {
		errMsg := "Internal Server Error"
		log.Println(errMsg + ": " + err.Error())
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(204)
}
