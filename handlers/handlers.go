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

func checkConversation(w http.ResponseWriter, db models.Datastore, id int64) (*models.Conversation, error) {
	conversation, err := db.GetConversation(id)
	if err != nil {
		errMsg := "Internal Server Error"
		log.Println(errMsg + ": " + err.Error())
		http.Error(w, errMsg, http.StatusInternalServerError)
		return nil, err
	}

	if conversation == nil {
		errMsg := fmt.Sprintf("Conversation %d does not exist", id)
		log.Println(errMsg)
		http.Error(w, "Conversation not found", http.StatusNotFound)
		return nil, nil
	}

	return conversation, nil
}

func checkMapping(w http.ResponseWriter, db models.Datastore, userID, convID int64) (*models.UserConversationMapping, error) {
	mapping, err := db.GetUserConversationMapping(userID, convID)
	if err != nil {
		errMsg := "Internal Server Error"
		log.Println(errMsg + ": " + err.Error())
		http.Error(w, errMsg, http.StatusInternalServerError)
		return nil, err
	}

	if mapping == nil {
		errMsg := fmt.Sprintf("User %d is not in conversation %d", userID, convID)
		log.Println(errMsg)
		http.Error(w, "Conversation not found", http.StatusNotFound)
		return nil, err
	}

	return mapping, nil
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

	conversation, err := checkConversation(w, env.DB, conversationID)
	if err != nil || conversation == nil {
		return
	}

	mapping, err := checkMapping(w, env.DB, userID, conversationID)
	if err != nil || mapping == nil {
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
		errMsg := "Invalid conversation ID"
		log.Println(errMsg + ": " + err.Error())
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	conversation, err := checkConversation(w, env.DB, conversationID)
	if err != nil || conversation == nil {
		return
	}

	mapping, err := checkMapping(w, env.DB, userID, conversationID)
	if err != nil || mapping == nil {
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

	conversation, err := checkConversation(w, env.DB, conversationID)
	if err != nil || conversation == nil {
		return
	}

	mapping, err := checkMapping(w, env.DB, userID, conversationID)
	if err != nil || mapping == nil {
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
