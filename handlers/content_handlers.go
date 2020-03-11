package handlers

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"

	"github.com/gorilla/mux"
)

// GetContentHandler gets a conversation's content
func (env *Env) GetContentHandler(w http.ResponseWriter, r *http.Request) {
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

	filePath := path.Join(contentDir, fmt.Sprintf("%d.html", conversationID))
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		errMsg := fmt.Sprintf("File for conversation %d does not exist", conversationID)
		log.Println(errMsg)
		http.Error(w, "File not found", http.StatusNotFound)
		return
	} else if err != nil {
		internalServerError(w, err)
		return
	}

	data, err := ioutil.ReadFile(filePath)
	w.Header().Add("Content-Type", "text/html")
	w.Write(data)
}
