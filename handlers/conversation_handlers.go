package handlers

import (
	"encoding/json"
	"ether/models"
	"ether/utils"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
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

	if reqConversation.Name == "" {
		errMsg := "Request body is missing mandatory field(s)"
		log.Println(errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	if reqConversation.Description == nil {
		reqConversation.Description = utils.StringPtr("")
	}
	if reqConversation.AvatarURL == nil {
		reqConversation.AvatarURL = utils.StringPtr("")
	}

	conversationID, err := env.DB.CreateConversation(reqConversation, userID)
	if err != nil {
		internalServerError(w, err)
		return
	}

	filePath := path.Join(contentDir, fmt.Sprintf("%d.html", conversationID))
	file, err := os.Create(filePath)
	if err != nil {
		internalServerError(w, err)
		return
	}
	file.Close()

	reqConversation.ID = conversationID
	location := fmt.Sprintf("%s/%d", r.URL.Path, conversationID)
	w.Header().Add("Location", location)
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(reqConversation)
}

// GetConversationsHandler returns all of a user's conversations
func (env *Env) GetConversationsHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(r.Header.Get("User-ID"), 10, 64)
	if err != nil {
		errMsg := "Invalid user ID"
		log.Println(errMsg + ": " + err.Error())
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	sort := r.URL.Query().Get("sort_by")
	if sort != "" && sort != "asc" && sort != "desc" {
		errMsg := "Invalid sorting keyword"
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	conversations, err := env.DB.GetConversations(userID, sort)
	if err != nil || conversations == nil {
		internalServerError(w, err)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	response := map[string][]models.Conversation{"conversations": conversations}
	json.NewEncoder(w).Encode(response)

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

	filePath := path.Join(contentDir, fmt.Sprintf("%d.html", conversationID))
	if err := os.Remove(filePath); err != nil {
		internalServerError(w, err)
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
		errMsg := `Request body must have one of "name", "description", or "avatar_url"`
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
