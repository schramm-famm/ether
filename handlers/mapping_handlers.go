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

	"github.com/go-sql-driver/mysql"
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
		errMsg := "Invalid conversation ID"
		log.Println(errMsg + ": " + err.Error())
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	conversation, err := env.getConversation(w, conversationID)
	if err != nil || conversation == nil {
		return
	}

	sessionMember, err := env.DB.GetUserConversationMapping(userID, conversationID)
	if err != nil {
		internalServerError(w, err)
		return
	}
	if sessionMember == nil {
		errMsg := fmt.Sprintf("User %d is not in conversation %d", userID, conversationID)
		log.Println(errMsg)
		http.Error(w, "Conversation not found", http.StatusNotFound)
		return
	}

	reqMember := &models.UserConversationMapping{}
	if err := parseMappingJSON(w, r.Body, reqMember); err != nil {
		return
	}

	if reqMember.UserID == 0 || reqMember.Role == "" {
		errMsg := "Request body is missing field(s)"
		log.Println(errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	if reqMember.Role != models.Admin && reqMember.Role != models.User ||
		sessionMember.Role != models.Owner && reqMember.Role != models.User {
		errMsg := "Invalid role value"
		log.Println(errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	// TODO: Validate that user is real

	reqMember.ConversationID = conversationID
	reqMember.Nickname = new(string)
	var pending bool = true
	reqMember.Pending = &pending
	reqMember.LastOpened = time.Now().Format("2006-01-02 15:04:05")
	err = env.DB.CreateUserConversationMapping(reqMember)
	if err != nil {
		mySQLErr, ok := err.(*mysql.MySQLError)
		if ok && mySQLErr.Number == 1062 {
			errMsg := fmt.Sprintf("User %d is already in conversation %d", reqMember.UserID, conversationID)
			log.Println(errMsg)
			http.Error(w, errMsg, http.StatusConflict)
		} else {
			internalServerError(w, err)
		}
		return
	}

	location := fmt.Sprintf("%s/%d", r.URL.Path, reqMember.UserID)
	w.Header().Add("Location", location)
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(reqMember)
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
	memberID, err := strconv.ParseInt(vars["user_id"], 10, 64)
	if err != nil {
		errMsg := "Invalid user ID"
		log.Println(errMsg + ": " + err.Error())
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	conversation, err := env.getConversation(w, conversationID)
	if err != nil || conversation == nil {
		return
	}

	sessionMember, err := env.DB.GetUserConversationMapping(userID, conversationID)
	if err != nil {
		internalServerError(w, err)
		return
	}
	if sessionMember == nil {
		errMsg := fmt.Sprintf("User %d is not in conversation %d", userID, conversationID)
		log.Println(errMsg)
		http.Error(w, "Conversation not found", http.StatusNotFound)
		return
	}

	member, err := env.DB.GetUserConversationMapping(memberID, conversationID)
	if err != nil {
		internalServerError(w, err)
		return
	}
	if member == nil {
		errMsg := fmt.Sprintf("User %d is not in conversation %d", userID, conversationID)
		log.Println(errMsg)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(member)
}

// GetMappingsHandler gets all users from a conversation
func (env *Env) GetMappingsHandler(w http.ResponseWriter, r *http.Request) {
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

	sessionMember, err := env.DB.GetUserConversationMapping(userID, conversationID)
	if err != nil {
		internalServerError(w, err)
		return
	}
	if sessionMember == nil {
		errMsg := fmt.Sprintf("User %d is not in conversation %d", userID, conversationID)
		log.Println(errMsg)
		http.Error(w, "Conversation not found", http.StatusNotFound)
		return
	}

	members, err := env.DB.GetUserConversationMappings(conversationID)
	if err != nil {
		internalServerError(w, err)
		return
	}
	memberList := &models.UserConversationMappingList{Users: members}

	// TODO: Get user names from karen?

	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(memberList)
}

// PatchMappingHandler updates a single user in a conversation
func (env *Env) PatchMappingHandler(w http.ResponseWriter, r *http.Request) {
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
		errMsg := "Invalid conversation ID"
		log.Println(errMsg + ": " + err.Error())
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}
	memberID, err := strconv.ParseInt(vars["user_id"], 10, 64)
	if err != nil {
		errMsg := "Invalid user ID"
		log.Println(errMsg + ": " + err.Error())
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}
	if userID == memberID {
		errMsg := "Cannot modify own role"
		log.Println(errMsg)
		http.Error(w, errMsg, http.StatusForbidden)
		return
	}

	conversation, err := env.getConversation(w, conversationID)
	if err != nil || conversation == nil {
		return
	}

	sessionMember, err := env.DB.GetUserConversationMapping(userID, conversationID)
	if err != nil {
		internalServerError(w, err)
		return
	}
	if sessionMember == nil {
		errMsg := fmt.Sprintf("User %d is not in conversation %d", userID, conversationID)
		log.Println(errMsg)
		http.Error(w, "Conversation not found", http.StatusNotFound)
		return
	}

	member, err := env.DB.GetUserConversationMapping(memberID, conversationID)
	if err != nil {
		internalServerError(w, err)
		return
	}
	if member == nil {
		errMsg := fmt.Sprintf("User %d is not in conversation %d", userID, conversationID)
		log.Println(errMsg)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	reqMember := &models.UserConversationMapping{}
	if err := parseMappingJSON(w, r.Body, reqMember); err != nil {
		return
	}

	if reqMember.Role == "" && reqMember.Nickname == nil {
		errMsg := "Request body is missing field(s)"
		log.Println(errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	if reqMember.Role != "" {
		if reqMember.Role != models.Admin && reqMember.Role != models.User {
			errMsg := "Invalid role value"
			log.Println(errMsg)
			http.Error(w, errMsg, http.StatusBadRequest)
			return
		}
		if sessionMember.Role != models.Owner {
			errMsg := fmt.Sprintf("User %d cannot modify roles in conversation %d", userID, conversationID)
			log.Println(errMsg)
			http.Error(w, "Forbidden from modifying roles", http.StatusForbidden)
			return
		}
	}

	newMember := member.Merge(reqMember)
	err = env.DB.UpdateUserConversationMapping(newMember)
	if err != nil {
		internalServerError(w, err)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newMember)
}

// DeleteMappingHandler deletes a single user from a conversation
func (env *Env) DeleteMappingHandler(w http.ResponseWriter, r *http.Request) {
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
	memberID, err := strconv.ParseInt(vars["user_id"], 10, 64)
	if err != nil {
		errMsg := "Invalid user ID"
		log.Println(errMsg + ": " + err.Error())
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}
	if userID == memberID {
		errMsg := "Cannot removed self from conversation"
		log.Println(errMsg)
		http.Error(w, errMsg, http.StatusForbidden)
		return
	}

	conversation, err := env.getConversation(w, conversationID)
	if err != nil || conversation == nil {
		return
	}

	sessionMember, err := env.DB.GetUserConversationMapping(userID, conversationID)
	if err != nil {
		internalServerError(w, err)
		return
	}
	if sessionMember == nil {
		errMsg := fmt.Sprintf("User %d is not in conversation %d", userID, conversationID)
		log.Println(errMsg)
		http.Error(w, "Conversation not found", http.StatusNotFound)
		return
	}

	member, err := env.DB.GetUserConversationMapping(memberID, conversationID)
	if err != nil {
		internalServerError(w, err)
		return
	}
	if member == nil {
		errMsg := fmt.Sprintf("User %d is not in conversation %d", userID, conversationID)
		log.Println(errMsg)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	res, err := sessionMember.Role.Compare(member.Role)
	if err != nil {
		internalServerError(w, err)
		return
	}
	if res != 1 {
		errMsg := fmt.Sprintf("User %d cannot remove user %d from conversation %d", userID, memberID, conversationID)
		log.Println(errMsg)
		http.Error(w, "Forbidden from removing this user", http.StatusForbidden)
		return
	}

	err = env.DB.DeleteUserConversationMapping(memberID, conversationID)
	if err != nil {
		internalServerError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
