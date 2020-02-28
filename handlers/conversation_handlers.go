package handlers

import (
	"encoding/json"
	"ether/models"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// PostConversationHandler creates a single new conversation
func (env *Env) PostConversationHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	userID, err := strconv.ParseInt(r.Header.Get("User-ID"), 10, 64)
	if err != nil {
		errMsg := "Invalid user ID"
		log.Println(errMsg + ": " + err.Error())
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	reqConversation := &models.Conversation{}
	if err := parseJSON(w, r.Body, reqConversation); err != nil {
		return
	}

	if reqConversation.Name == "" || reqConversation.Description == nil {
		errMsg := "Request body is missing field(s)"
		log.Println(errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	conversationID, err := env.DB.CreateConversation(reqConversation, userID)
	if err != nil {
		internalServerError(w, err)
		return
	}

	// TODO: Create HTML file

	reqConversation.ID = conversationID
	location := fmt.Sprintf("%s/%d", r.URL.Path, conversationID)
	w.Header().Add("Location", location)
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(reqConversation)
}

// GetConversationHandler gets a single conversation
func (env *Env) GetConversationHandler(w http.ResponseWriter, r *http.Request) {
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

	conversation, err := env.getConversation(w, conversationID)
	if err != nil || conversation == nil {
		return
	}

	sessionMember, err := env.getMapping(w, userID, conversationID, "Conversation not found")
	if err != nil || sessionMember == nil {
		return
	}

	// TODO: Get conversation HTML
	// TODO: Get conversation picture

	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(conversation)
}

// DeleteConversationHandler deletes a single conversation
func (env *Env) DeleteConversationHandler(w http.ResponseWriter, r *http.Request) {
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

	conversation, err := env.getConversation(w, conversationID)
	if err != nil || conversation == nil {
		return
	}

	sessionMember, err := env.getMapping(w, userID, conversationID, "Conversation not found")
	if err != nil || sessionMember == nil {
		return
	}

	if sessionMember.Role != models.Owner {
		errMsg := fmt.Sprintf("User %d is not an Owner of conversation %d and cannot delete it", userID, conversationID)
		log.Println(errMsg)
		http.Error(w, "Forbidden from deleting conversation", http.StatusForbidden)
		return
	}

	err = env.DB.DeleteConversation(conversationID)
	if err != nil {
		internalServerError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// PatchConversationHandler updates a single conversation
func (env *Env) PatchConversationHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	userID, err := strconv.ParseInt(r.Header.Get("User-ID"), 10, 64)
	if err != nil {
		errMsg := "Invalid user ID"
		log.Println(errMsg + ": " + err.Error())
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	reqConversation := &models.Conversation{}
	if err := parseJSON(w, r.Body, reqConversation); err != nil {
		return
	}

	if reqConversation.Name == "" && reqConversation.Description == nil && reqConversation.AvatarURL == nil {
		errMsg := `Request body has neither "name" nor "description"`
		log.Println(errMsg)
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

	conversation, err := env.getConversation(w, conversationID)
	if err != nil || conversation == nil {
		return
	}

	sessionMember, err := env.getMapping(w, userID, conversationID, "Conversation not found")
	if err != nil || sessionMember == nil {
		return
	}

	if *sessionMember.Pending {
		errMsg := "Cannot modify conversation while invitation is pending"
		log.Println(errMsg)
		http.Error(w, errMsg, http.StatusForbidden)
		return
	}

	newConversation := conversation.Merge(reqConversation)

	err = env.DB.UpdateConversation(newConversation)
	if err != nil {
		internalServerError(w, err)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newConversation)
}
