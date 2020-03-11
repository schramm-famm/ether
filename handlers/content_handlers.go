package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"
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

	if *sessionMember.Pending {
		errMsg := "Cannot get conversation content while invitation is pending"
		log.Println(errMsg)
		http.Error(w, errMsg, http.StatusForbidden)
		return
	}

	data, err := env.Directory.ReadFile(conversationID)
	if os.IsNotExist(err) {
		errMsg := fmt.Sprintf("File for conversation %d does not exist", conversationID)
		log.Println(errMsg)
		http.Error(w, "File not found", http.StatusNotFound)
		return
	} else if err != nil {
		internalServerError(w, err)
		return
	}

	w.Header().Add("Content-Type", "text/html")
	w.Write(data)
}
