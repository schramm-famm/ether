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
	"time"

	"github.com/gorilla/mux"
)

func parseMappingJSON(w http.ResponseWriter, body io.ReadCloser, bodyObj *models.UserConversationMapping) error {
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

// PostMappingHandler adds a single user to a conversation
func (env *Env) PostMappingHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	userID, err := strconv.ParseInt(r.Header.Get("User-ID"), 10, 64)
	if err != nil {
		errMsg := "Invalid user ID"
		log.Println(errMsg + ": " + err.Error())
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	conversationID, err := strconv.ParseInt(vars["conversation_id"], 10, 64)
	if err != nil {
		errMsg := "Invalid ID"
		log.Println(errMsg + ": " + err.Error())
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	conversation, err := env.getConversation(w, conversationID)
	if err != nil || conversation == nil {
		return
	}

	mapping, err := env.DB.GetUserConversationMapping(userID, conversationID)
	if err != nil {
		internalServerError(w, err)
		return
	}
	if mapping == nil {
		errMsg := fmt.Sprintf("User %d is not in conversation %d", userID, conversationID)
		log.Println(errMsg)
		http.Error(w, "Conversation not found", http.StatusNotFound)
		return
	}

	reqMapping := &models.UserConversationMapping{}
	if err := parseMappingJSON(w, r.Body, reqMapping); err != nil {
		return
	}

	if reqMapping.Role == "" {
		errMsg := "Request body is missing field(s)"
		log.Println(errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	if reqMapping.Role != models.Admin && reqMapping.Role != models.User {
		errMsg := "Invalid role value"
		log.Println(errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	// TODO: Validate that user is real

	reqMapping.ConversationID = conversationID
	reqMapping.Nickname = new(string)
	var pending bool = true
	reqMapping.Pending = &pending
	reqMapping.LastOpened = time.Now().Format("2006-01-02 15:04:05")
	err = env.DB.CreateUserConversationMapping(reqMapping)
	if err != nil {
		internalServerError(w, err)
		return
	}

	location := fmt.Sprintf("%s/%d", r.URL.Path, reqMapping.UserID)
	w.Header().Add("Location", location)
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(reqMapping)
}

// GetMappingHandler gets a single user from a conversation
func (env *Env) GetMappingHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(r.Header.Get("User-ID"), 10, 64)
	if err != nil {
		errMsg := "Invalid user ID"
		log.Println(errMsg + ": " + err.Error())
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	conversationID, err := strconv.ParseInt(vars["conversation_id"], 10, 64)
	if err != nil {
		errMsg := "Invalid conversation ID"
		log.Println(errMsg + ": " + err.Error())
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}
	mappedUserID, err := strconv.ParseInt(vars["user_id"], 10, 64)
	if err != nil {
		errMsg := "Invalid user ID"
		log.Println(errMsg + ": " + err.Error())
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	mapping, err := env.getMapping(w, userID, conversationID)
	if err != nil {
		return
	}
	if mapping == nil {
		errMsg := fmt.Sprintf("User %d is not in conversation %d", userID, conversationID)
		log.Println(errMsg)
		http.Error(w, "Conversation not found", http.StatusNotFound)
		return
	}

	mapping, err = env.getMapping(w, mappedUserID, conversationID)
	if err != nil {
		return
	}
	if mapping == nil {
		errMsg := fmt.Sprintf("User %d is not in conversation %d", userID, conversationID)
		log.Println(errMsg)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mapping)
}

// GetMappingsHandler gets all users from a conversation
func (env *Env) GetMappingsHandler(w http.ResponseWriter, r *http.Request) {}

// PatchMappingHandler updates a single user in a conversation
func (env *Env) PatchMappingHandler(w http.ResponseWriter, r *http.Request) {}

// DeleteMappingHandler deletes a single user from a conversation
func (env *Env) DeleteMappingHandler(w http.ResponseWriter, r *http.Request) {}
