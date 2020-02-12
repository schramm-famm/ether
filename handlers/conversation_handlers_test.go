package handlers

import (
	"bytes"
	"encoding/json"
	"ether/models"
	"ether/utils"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"

	"github.com/gorilla/mux"
)

func TestPostConversationsHandler(t *testing.T) {
	tests := []struct {
		Name       string
		StatusCode int
		ReqBody    interface{}
		ResBody    *models.Conversation
		Location   string
	}{
		{
			Name:       "Successful conversation creation",
			StatusCode: http.StatusCreated,
			ReqBody: map[string]interface{}{
				"name":        "test_name",
				"description": "test_desc",
			},
			ResBody: &models.Conversation{
				ID:          1,
				Name:        "test_name",
				Description: utils.StringPtr("test_desc"),
			},
			Location: "/ether/v1/conversations/1",
		},
		{
			Name:       "Successful conversation creation (empty description)",
			StatusCode: http.StatusCreated,
			ReqBody: map[string]interface{}{
				"name":        "test_name",
				"description": "",
			},
			ResBody: &models.Conversation{
				ID:          1,
				Name:        "test_name",
				Description: utils.StringPtr(""),
			},
			Location: "/ether/v1/conversations/1",
		},
		{
			Name:       "Failed conversation creation (empty name)",
			StatusCode: http.StatusBadRequest,
			ReqBody: map[string]interface{}{
				"name":        "",
				"description": "test_desc",
			},
		},
		{
			Name:       "Failed conversation creation (wrong name type)",
			StatusCode: http.StatusBadRequest,
			ReqBody: map[string]interface{}{
				"name":        13,
				"description": "test_desc",
			},
		},
		{
			Name:       "Failed conversation creation (wrong description type)",
			StatusCode: http.StatusBadRequest,
			ReqBody: map[string]interface{}{
				"name":        "test_name",
				"description": 13,
			},
		},
		{
			Name:       "Failed conversation creation (empty JSON)",
			StatusCode: http.StatusBadRequest,
			ReqBody:    map[string]interface{}{},
		},
	}

	var userID int64 = 1
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			reqBody, _ := json.Marshal(test.ReqBody)
			r := httptest.NewRequest("POST", "/ether/v1/conversations", bytes.NewReader(reqBody))
			r.Header.Set("User-ID", strconv.FormatInt(userID, 10))
			w := httptest.NewRecorder()

			mDB := models.NewMockDB(nil, nil, nil)

			env := &Env{DB: mDB}
			env.PostConversationHandler(w, r)

			if w.Code != test.StatusCode {
				t.Errorf("Response has incorrect status code, expected status code %d, got %d", test.StatusCode, w.Code)
			}

			if w.Code == http.StatusCreated {
				// Validate HTTP response content
				resBody := models.Conversation{}
				_ = json.NewDecoder(w.Body).Decode(&resBody)
				if !reflect.DeepEqual(*test.ResBody, resBody) {
					t.Errorf("Response has incorrect body, expected %+v, got %+v", *test.ResBody, resBody)
				}
				if test.Location != w.Header().Get("Location") {
					t.Errorf(
						`Response has incorrect "Location" header, expected %s, got %s`,
						test.Location,
						w.Header().Get("Location"),
					)
				}

				// Validate DB function calls
				if !reflect.DeepEqual(*test.ResBody, *mDB.Conversations[test.ResBody.ID]) {
					t.Errorf(
						"Used incorrect conversation, expected %+v, got %+v",
						*test.ResBody,
						*mDB.Conversations[test.ResBody.ID],
					)
				}
				if mDB.GetMapping(userID, test.ResBody.ID) == nil {
					t.Errorf("Used incorrect user ID as owner, expected %d", userID)
				}
			}
		})
	}
}

func TestGetConversationsHandler(t *testing.T) {
	tests := []struct {
		Name       string
		StatusCode int
		ResBody    *models.Conversation
		Mapping    *models.UserConversationMapping
	}{
		{
			Name:       "Successful conversation retrieval",
			StatusCode: http.StatusOK,
			ResBody: &models.Conversation{
				ID:          1,
				Name:        "test_name",
				Description: utils.StringPtr("test_desc"),
			},
			Mapping: &models.UserConversationMapping{
				UserID:         1,
				ConversationID: 1,
				Role:           "owner",
				Nickname:       utils.StringPtr(""),
				Pending:        utils.BoolPtr(false),
				LastOpened:     "2006-01-02 15:04:05",
			},
		},
		{
			Name:       "Failed conversation retrieval (conversation does not exist)",
			StatusCode: http.StatusNotFound,
		},
		{
			Name:       "Failed conversation retrieval (user not in conversation)",
			StatusCode: http.StatusNotFound,
			ResBody: &models.Conversation{
				ID:          1,
				Name:        "test_name",
				Description: utils.StringPtr("test_desc"),
			},
		},
	}

	var userID int64 = 1
	var conversationID int64 = 1
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/ether/v1/conversations/1", nil)
			r.Header.Set("User-ID", strconv.FormatInt(userID, 10))
			r = mux.SetURLVars(r, map[string]string{
				"conversation_id": strconv.FormatInt(conversationID, 10),
			})
			w := httptest.NewRecorder()

			mDB := models.NewMockDB(
				[]*models.Conversation{test.ResBody},
				[]*models.UserConversationMapping{test.Mapping},
				nil,
			)

			env := &Env{DB: mDB}
			env.GetConversationHandler(w, r)

			if w.Code != test.StatusCode {
				t.Errorf("Response has incorrect status code, expected status code %d, got %d", test.StatusCode, w.Code)
			}

			if w.Code == http.StatusOK {
				// Validate HTTP response content
				resBody := models.Conversation{}
				_ = json.NewDecoder(w.Body).Decode(&resBody)
				if !reflect.DeepEqual(*test.ResBody, resBody) {
					t.Errorf("Response has incorrect body, expected %+v, got %+v", *test.ResBody, resBody)
				}
			}
		})
	}
}

func TestPatchConversationsHandler(t *testing.T) {
	tests := []struct {
		Name                string
		StatusCode          int
		ReqBody             interface{}
		ResBody             *models.Conversation
		Mapping             *models.UserConversationMapping
		InitialConversation bool
	}{
		{
			Name:       "Successful conversation modification",
			StatusCode: http.StatusOK,
			ReqBody: map[string]interface{}{
				"name":        "test_new_name",
				"description": "test_new_desc",
			},
			ResBody: &models.Conversation{
				ID:          1,
				Name:        "test_new_name",
				Description: utils.StringPtr("test_new_desc"),
			},
			Mapping: &models.UserConversationMapping{
				UserID:         1,
				ConversationID: 1,
				Role:           "owner",
				Nickname:       utils.StringPtr(""),
				Pending:        utils.BoolPtr(false),
				LastOpened:     "2006-01-02 15:04:05",
			},
			InitialConversation: true,
		},
		{
			Name:       "Successful conversation modification (just name)",
			StatusCode: http.StatusOK,
			ReqBody: map[string]interface{}{
				"name": "test_new_name",
			},
			ResBody: &models.Conversation{
				ID:          1,
				Name:        "test_new_name",
				Description: utils.StringPtr("test_desc"),
			},
			Mapping: &models.UserConversationMapping{
				UserID:         1,
				ConversationID: 1,
				Role:           "owner",
				Nickname:       utils.StringPtr(""),
				Pending:        utils.BoolPtr(false),
				LastOpened:     "2006-01-02 15:04:05",
			},
			InitialConversation: true,
		},
		{
			Name:       "Successful conversation modification (just description)",
			StatusCode: http.StatusOK,
			ReqBody: map[string]interface{}{
				"description": "test_new_desc",
			},
			ResBody: &models.Conversation{
				ID:          1,
				Name:        "test_name",
				Description: utils.StringPtr("test_new_desc"),
			},
			Mapping: &models.UserConversationMapping{
				UserID:         1,
				ConversationID: 1,
				Role:           "owner",
				Nickname:       utils.StringPtr(""),
				Pending:        utils.BoolPtr(false),
				LastOpened:     "2006-01-02 15:04:05",
			},
			InitialConversation: true,
		},
		{
			Name:       "Failed conversation modification (conversation does not exist)",
			StatusCode: http.StatusNotFound,
			ReqBody: map[string]interface{}{
				"name": "test_new_name",
			},
			InitialConversation: false,
		},
		{
			Name:       "Failed conversation modification (user not in conversation)",
			StatusCode: http.StatusNotFound,
			ReqBody: map[string]interface{}{
				"name": "test_new_name",
			},
			InitialConversation: true,
		},
		{
			Name:       "Successful conversation modification (pending invitation)",
			StatusCode: http.StatusForbidden,
			ReqBody: map[string]interface{}{
				"name": "test_new_name",
			},
			Mapping: &models.UserConversationMapping{
				UserID:         1,
				ConversationID: 1,
				Role:           "admin",
				Nickname:       utils.StringPtr(""),
				Pending:        utils.BoolPtr(true),
				LastOpened:     "2006-01-02 15:04:05",
			},
			InitialConversation: true,
		},
		{
			Name:       "Failed conversation modification (wrong name type)",
			StatusCode: http.StatusBadRequest,
			ReqBody: map[string]interface{}{
				"name": 13,
			},
			InitialConversation: true,
		},
		{
			Name:       "Failed conversation modification (wrong description type)",
			StatusCode: http.StatusBadRequest,
			ReqBody: map[string]interface{}{
				"description": 13,
			},
			InitialConversation: true,
		},
		{
			Name:                "Failed conversation modification (empty JSON)",
			StatusCode:          http.StatusBadRequest,
			ReqBody:             map[string]interface{}{},
			InitialConversation: true,
		},
	}

	var userID int64 = 1
	var conversationID int64 = 1
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			reqBody, _ := json.Marshal(test.ReqBody)
			r := httptest.NewRequest("GET", "/ether/v1/conversations/1", bytes.NewReader(reqBody))
			r.Header.Set("User-ID", strconv.FormatInt(userID, 10))
			r = mux.SetURLVars(r, map[string]string{
				"conversation_id": strconv.FormatInt(conversationID, 10),
			})
			w := httptest.NewRecorder()

			mockConversation := make([]*models.Conversation, 0)
			if test.InitialConversation {
				mockConversation = append(mockConversation, &models.Conversation{
					ID:          conversationID,
					Name:        "test_name",
					Description: utils.StringPtr("test_desc"),
				})
			}
			mDB := models.NewMockDB(
				mockConversation,
				[]*models.UserConversationMapping{test.Mapping},
				nil,
			)

			env := &Env{DB: mDB}
			env.PatchConversationHandler(w, r)

			if w.Code != test.StatusCode {
				t.Errorf("Response has incorrect status code, expected status code %d, got %d", test.StatusCode, w.Code)
			}

			if w.Code == http.StatusOK {
				// Validate HTTP response content
				resBody := models.Conversation{}
				_ = json.NewDecoder(w.Body).Decode(&resBody)
				if !reflect.DeepEqual(*test.ResBody, resBody) {
					t.Errorf("Response has incorrect body, expected %+v, got %+v", *test.ResBody, resBody)
				}

				// Validate DB function calls
				if !reflect.DeepEqual(*test.ResBody, *mDB.Conversations[conversationID]) {
					t.Errorf(
						"Used incorrect conversation patch, expected %+v, got %+v",
						*test.ResBody,
						*mDB.Conversations[conversationID],
					)
				}
			}
		})
	}
}

func TestDeleteConversationsHandler(t *testing.T) {
	tests := []struct {
		Name                string
		StatusCode          int
		Mapping             *models.UserConversationMapping
		InitialConversation bool
	}{
		{
			Name:       "Successful conversation deletion",
			StatusCode: http.StatusNoContent,
			Mapping: &models.UserConversationMapping{
				UserID:         1,
				ConversationID: 1,
				Role:           "owner",
				Nickname:       utils.StringPtr(""),
				Pending:        utils.BoolPtr(false),
				LastOpened:     "2006-01-02 15:04:05",
			},
			InitialConversation: true,
		},
		{
			Name:                "Failed conversation deletion (conversation does not exist)",
			StatusCode:          http.StatusNotFound,
			InitialConversation: false,
		},
		{
			Name:                "Failed conversation deletion (user not in conversation)",
			StatusCode:          http.StatusNotFound,
			InitialConversation: true,
		},
		{
			Name:       "Failed conversation deletion (role is user)",
			StatusCode: http.StatusForbidden,
			Mapping: &models.UserConversationMapping{
				UserID:         1,
				ConversationID: 1,
				Role:           "user",
				Nickname:       utils.StringPtr(""),
				Pending:        utils.BoolPtr(false),
				LastOpened:     "2006-01-02 15:04:05",
			},
			InitialConversation: true,
		},
		{
			Name:       "Failed conversation deletion (role is admin)",
			StatusCode: http.StatusForbidden,
			Mapping: &models.UserConversationMapping{
				UserID:         1,
				ConversationID: 1,
				Role:           "admin",
				Nickname:       utils.StringPtr(""),
				Pending:        utils.BoolPtr(false),
				LastOpened:     "2006-01-02 15:04:05",
			},
			InitialConversation: true,
		},
	}

	var userID int64 = 1
	var conversationID int64 = 1
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			r := httptest.NewRequest("DELETE", "/ether/v1/conversations/1", nil)
			r.Header.Set("User-ID", strconv.FormatInt(userID, 10))
			r = mux.SetURLVars(r, map[string]string{
				"conversation_id": strconv.FormatInt(conversationID, 10),
			})
			w := httptest.NewRecorder()

			mockConversation := make([]*models.Conversation, 0)
			if test.InitialConversation {
				mockConversation = append(mockConversation, &models.Conversation{
					ID:          conversationID,
					Name:        "test_name",
					Description: utils.StringPtr("test_desc"),
				})
			}
			mDB := models.NewMockDB(
				mockConversation,
				[]*models.UserConversationMapping{test.Mapping},
				nil,
			)

			env := &Env{DB: mDB}
			env.DeleteConversationHandler(w, r)

			if w.Code != test.StatusCode {
				t.Errorf("Response has incorrect status code, expected status code %d, got %d", test.StatusCode, w.Code)
			}
		})
	}
}
