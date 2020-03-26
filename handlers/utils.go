package handlers

import (
	"encoding/json"
	"ether/filesystem"
	"ether/models"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

// Env represents all application-level items that are needed by HTTP handlers.
type Env struct {
	DB        models.Datastore
	Directory *filesystem.Directory
	Client    *http.Client
	KarenHost string
}

func internalServerError(w http.ResponseWriter, err error) {
	errMsg := "Internal Server Error"
	log.Println(errMsg + ": " + err.Error())
	http.Error(w, errMsg, http.StatusInternalServerError)
}

func parseJSON(w http.ResponseWriter, body io.ReadCloser, bodyObj interface{}) error {
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
		internalServerError(w, err)
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

func (env *Env) getMapping(
	w http.ResponseWriter,
	userID int64,
	conversationID int64,
	httpNotFoundMsg string,
) (*models.UserConversationMapping, error) {
	mapping, err := env.DB.GetUserConversationMapping(userID, conversationID)
	if err != nil {
		internalServerError(w, err)
		return nil, err
	}

	if mapping == nil {
		errMsg := fmt.Sprintf("User %d is not in conversation %d", userID, conversationID)
		log.Println(errMsg)
		http.Error(w, httpNotFoundMsg, http.StatusNotFound)
		return nil, nil
	}

	return mapping, nil
}
