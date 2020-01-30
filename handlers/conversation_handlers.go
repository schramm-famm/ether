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

func (env *Env) getConversation(w http.ResponseWriter, id int64) (*models.Conversation, error) {
	conversation, err := env.DB.GetConversation(id)
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

func (env *Env) getMapping(w http.ResponseWriter, userID, convID int64) (*models.UserConversationMapping, error) {
	mapping, err := env.DB.GetUserConversationMapping(userID, convID)
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
	if err := parseReqBody(w, r.Body, reqConversation); err != nil {
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
		errMsg := "Internal Server Error"
		log.Println(errMsg + ": " + err.Error())
		http.Error(w, errMsg, http.StatusInternalServerError)
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
		errMsg := "Invalid ID"
		log.Println(errMsg + ": " + err.Error())
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	conversation, err := env.getConversation(w, conversationID)
	if err != nil || conversation == nil {
		return
	}

	mapping, err := env.getMapping(w, userID, conversationID)
	if err != nil || mapping == nil {
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

	mapping, err := env.getMapping(w, userID, conversationID)
	if err != nil || mapping == nil {
		return
	}

	if mapping.Role != models.Owner {
		errMsg := fmt.Sprintf("User %d is not an Owner of conversation %d and cannot delete it", userID, conversationID)
		log.Println(errMsg)
		http.Error(w, "Forbidden from deleting conversation", http.StatusForbidden)
		return
	}

	err = env.DB.DeleteConversation(conversationID)
	if err != nil {
		errMsg := "Internal Server Error"
		log.Println(errMsg + ": " + err.Error())
		http.Error(w, errMsg, http.StatusInternalServerError)
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
	if err := parseReqBody(w, r.Body, reqConversation); err != nil {
		return
	}

	if reqConversation.Name == "" && reqConversation.Description == nil {
		errMsg := `Request body has neither "name" nor "description"`
		log.Println(errMsg)
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

	mapping, err := env.getMapping(w, userID, conversationID)
	if err != nil || mapping == nil {
		return
	}

	newConversation := conversation.Merge(reqConversation)

	err = env.DB.UpdateConversation(newConversation)
	if err != nil {
		errMsg := "Internal Server Error"
		log.Println(errMsg + ": " + err.Error())
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newConversation)
}
