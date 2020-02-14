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

	"github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

func TestPostMappingHandler(t *testing.T) {
	tests := []struct {
		Name          string
		StatusCode    int
		ReqBody       interface{}
		ResBody       *models.UserConversationMapping
		Conversation  *models.Conversation
		SessionMember *models.UserConversationMapping
		Location      string
		Error         *mysql.MySQLError
	}{
		{
			Name:       "Successful member creation (owner creates user)",
			StatusCode: http.StatusCreated,
			ReqBody: map[string]interface{}{
				"user_id": 1,
				"role":    "user",
			},
			ResBody: &models.UserConversationMapping{
				UserID:         1,
				ConversationID: 11,
				Role:           "user",
				Nickname:       utils.StringPtr(""),
				Pending:        utils.BoolPtr(true),
			},
			Conversation: &models.Conversation{
				ID:          11,
				Name:        "testname",
				Description: utils.StringPtr("testdesc"),
			},
			SessionMember: &models.UserConversationMapping{
				UserID:         1337,
				ConversationID: 11,
				Role:           "owner",
				Nickname:       utils.StringPtr("testowner"),
				Pending:        utils.BoolPtr(false),
			},
			Location: "/ether/v1/conversations/11/users/1",
		},
		{
			Name:       "Successful member creation (owner creates admin)",
			StatusCode: http.StatusCreated,
			ReqBody: map[string]interface{}{
				"user_id": 1,
				"role":    "admin",
			},
			ResBody: &models.UserConversationMapping{
				UserID:         1,
				ConversationID: 11,
				Role:           "admin",
				Nickname:       utils.StringPtr(""),
				Pending:        utils.BoolPtr(true),
			},
			Conversation: &models.Conversation{
				ID:          11,
				Name:        "testname",
				Description: utils.StringPtr("testdesc"),
			},
			SessionMember: &models.UserConversationMapping{
				UserID:         1337,
				ConversationID: 11,
				Role:           "owner",
				Nickname:       utils.StringPtr("testowner"),
				Pending:        utils.BoolPtr(false),
			},
			Location: "/ether/v1/conversations/11/users/1",
		},
		{
			Name:       "Successful member creation (admin creates user)",
			StatusCode: http.StatusCreated,
			ReqBody: map[string]interface{}{
				"user_id": 1,
				"role":    "user",
			},
			ResBody: &models.UserConversationMapping{
				UserID:         1,
				ConversationID: 11,
				Role:           "user",
				Nickname:       utils.StringPtr(""),
				Pending:        utils.BoolPtr(true),
			},
			Conversation: &models.Conversation{
				ID:          11,
				Name:        "testname",
				Description: utils.StringPtr("testdesc"),
			},
			SessionMember: &models.UserConversationMapping{
				UserID:         1337,
				ConversationID: 11,
				Role:           "admin",
				Nickname:       utils.StringPtr("testadmin"),
				Pending:        utils.BoolPtr(false),
			},
			Location: "/ether/v1/conversations/11/users/1",
		},
		{
			Name:       "Successful member creation (user creates user)",
			StatusCode: http.StatusCreated,
			ReqBody: map[string]interface{}{
				"user_id": 1,
				"role":    "user",
			},
			ResBody: &models.UserConversationMapping{
				UserID:         1,
				ConversationID: 11,
				Role:           "user",
				Nickname:       utils.StringPtr(""),
				Pending:        utils.BoolPtr(true),
			},
			Conversation: &models.Conversation{
				ID:          11,
				Name:        "testname",
				Description: utils.StringPtr("testdesc"),
			},
			SessionMember: &models.UserConversationMapping{
				UserID:         1337,
				ConversationID: 11,
				Role:           "user",
				Nickname:       utils.StringPtr("testuser"),
				Pending:        utils.BoolPtr(false),
			},
			Location: "/ether/v1/conversations/11/users/1",
		},
		{
			Name:       "Failed member creation (conversation does not exist)",
			StatusCode: http.StatusNotFound,
			ReqBody: map[string]interface{}{
				"user_id": 1,
				"role":    "foobar",
			},
		},
		{
			Name:       "Failed member creation (user not in conversation)",
			StatusCode: http.StatusNotFound,
			ReqBody: map[string]interface{}{
				"user_id": 1,
				"role":    "user",
			},
			Conversation: &models.Conversation{
				ID:          11,
				Name:        "testname",
				Description: utils.StringPtr("testdesc"),
			},
		},
		{
			Name:       "Failed member creation (owner creates owner)",
			StatusCode: http.StatusBadRequest,
			ReqBody: map[string]interface{}{
				"user_id": 1,
				"role":    "owner",
			},
			Conversation: &models.Conversation{
				ID:          11,
				Name:        "testname",
				Description: utils.StringPtr("testdesc"),
			},
			SessionMember: &models.UserConversationMapping{
				UserID:         1337,
				ConversationID: 11,
				Role:           "owner",
				Nickname:       utils.StringPtr("testowner"),
				Pending:        utils.BoolPtr(false),
			},
		},
		{
			Name:       "Failed member creation (admin creates admin)",
			StatusCode: http.StatusBadRequest,
			ReqBody: map[string]interface{}{
				"user_id": 1,
				"role":    "admin",
			},
			Conversation: &models.Conversation{
				ID:          11,
				Name:        "testname",
				Description: utils.StringPtr("testdesc"),
			},
			SessionMember: &models.UserConversationMapping{
				UserID:         1337,
				ConversationID: 11,
				Role:           "admin",
				Nickname:       utils.StringPtr("testadmin"),
				Pending:        utils.BoolPtr(false),
			},
		},
		{
			Name:       "Failed member creation (pending user creates user)",
			StatusCode: http.StatusForbidden,
			ReqBody: map[string]interface{}{
				"user_id": 1,
				"role":    "user",
			},
			Conversation: &models.Conversation{
				ID:          11,
				Name:        "testname",
				Description: utils.StringPtr("testdesc"),
			},
			SessionMember: &models.UserConversationMapping{
				UserID:         1337,
				ConversationID: 11,
				Role:           "user",
				Nickname:       utils.StringPtr("testuser"),
				Pending:        utils.BoolPtr(true),
			},
		},
		{
			Name:       "Failed member creation (empty role)",
			StatusCode: http.StatusBadRequest,
			ReqBody: map[string]interface{}{
				"user_id": 1,
				"role":    "owner",
			},
			Conversation: &models.Conversation{
				ID:          11,
				Name:        "testname",
				Description: utils.StringPtr("testdesc"),
			},
			SessionMember: &models.UserConversationMapping{
				UserID:         1337,
				ConversationID: 11,
				Role:           "owner",
				Nickname:       utils.StringPtr("testowner"),
				Pending:        utils.BoolPtr(false),
			},
		},
		{
			Name:       "Failed member creation (invalid role string)",
			StatusCode: http.StatusBadRequest,
			ReqBody: map[string]interface{}{
				"user_id": 1,
				"role":    "foobar",
			},
			Conversation: &models.Conversation{
				ID:          11,
				Name:        "testname",
				Description: utils.StringPtr("testdesc"),
			},
			SessionMember: &models.UserConversationMapping{
				UserID:         1337,
				ConversationID: 11,
				Role:           "owner",
				Nickname:       utils.StringPtr("testowner"),
				Pending:        utils.BoolPtr(false),
			},
		},
		{
			Name:       "Failed member creation (member already in conversation)",
			StatusCode: http.StatusConflict,
			ReqBody: map[string]interface{}{
				"user_id": 1,
				"role":    "user",
			},
			SessionMember: &models.UserConversationMapping{
				UserID:         1337,
				ConversationID: 11,
				Role:           "owner",
				Nickname:       utils.StringPtr("testowner"),
				Pending:        utils.BoolPtr(false),
			},
			Conversation: &models.Conversation{
				ID:          11,
				Name:        "testname",
				Description: utils.StringPtr("testdesc"),
			},
			Error: &mysql.MySQLError{Number: 1062, Message: ""},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			var userID int64 = 1337
			var memberID int64 = 1
			var conversationID int64 = 11

			reqBody, _ := json.Marshal(test.ReqBody)
			r := httptest.NewRequest("POST", "/ether/v1/conversations/11/users", bytes.NewReader(reqBody))
			r.Header.Set("User-ID", strconv.FormatInt(userID, 10))
			r = mux.SetURLVars(r, map[string]string{
				"conversation_id": strconv.FormatInt(conversationID, 10),
			})
			w := httptest.NewRecorder()

			var errList []error = nil
			if test.Error != nil {
				errList = make([]error, 3)
				errList[2] = test.Error
			}
			mDB := models.NewMockDB(
				[]*models.Conversation{test.Conversation},
				[]*models.UserConversationMapping{test.SessionMember},
				errList,
			)

			env := &Env{DB: mDB}
			env.PostMappingHandler(w, r)

			if w.Code != test.StatusCode {
				t.Errorf("Response has incorrect status code, expected status code %d, got %d", test.StatusCode, w.Code)
				return
			}

			if w.Code == http.StatusCreated {
				// Validate HTTP response content
				resBody := models.UserConversationMapping{}
				_ = json.NewDecoder(w.Body).Decode(&resBody)
				test.ResBody.LastOpened = resBody.LastOpened
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
				if !reflect.DeepEqual(*test.ResBody, *mDB.GetMapping(memberID, conversationID)) {
					t.Errorf(
						"Used incorrect mapping, expected %+v, got %+v",
						*test.ResBody,
						mDB.GetMapping(memberID, conversationID),
					)
				}
			}
		})
	}
}

func TestGetMappingHandler(t *testing.T) {
	tests := []struct {
		Name          string
		StatusCode    int
		ResBody       *models.UserConversationMapping
		Conversation  *models.Conversation
		SessionMember *models.UserConversationMapping
	}{
		{
			Name:       "Successful member retrieval (other member)",
			StatusCode: http.StatusOK,
			ResBody: &models.UserConversationMapping{
				UserID:         1,
				ConversationID: 11,
				Role:           "user",
				Nickname:       utils.StringPtr(""),
				Pending:        utils.BoolPtr(true),
			},
			Conversation: &models.Conversation{
				ID:          11,
				Name:        "testname",
				Description: utils.StringPtr("testdesc"),
			},
			SessionMember: &models.UserConversationMapping{
				UserID:         1337,
				ConversationID: 11,
				Role:           "owner",
				Nickname:       utils.StringPtr("testowner"),
				Pending:        utils.BoolPtr(false),
			},
		},
		{
			Name:       "Successful member retrieval (self)",
			StatusCode: http.StatusOK,
			ResBody: &models.UserConversationMapping{
				UserID:         1,
				ConversationID: 11,
				Role:           "owner",
				Nickname:       utils.StringPtr("testowner"),
				Pending:        utils.BoolPtr(false),
			},
			Conversation: &models.Conversation{
				ID:          11,
				Name:        "testname",
				Description: utils.StringPtr("testdesc"),
			},
			SessionMember: &models.UserConversationMapping{
				UserID:         1,
				ConversationID: 11,
				Role:           "owner",
				Nickname:       utils.StringPtr("testowner"),
				Pending:        utils.BoolPtr(false),
			},
		},
		{
			Name:       "Failed member retrieval (member does not exist)",
			StatusCode: http.StatusNotFound,
			Conversation: &models.Conversation{
				ID:          11,
				Name:        "testname",
				Description: utils.StringPtr("testdesc"),
			},
			SessionMember: &models.UserConversationMapping{
				UserID:         1337,
				ConversationID: 11,
				Role:           "owner",
				Nickname:       utils.StringPtr("testowner"),
				Pending:        utils.BoolPtr(false),
			},
		},
		{
			Name:       "Failed member retrieval (conversation does not exist)",
			StatusCode: http.StatusNotFound,
		},
		{
			Name:       "Failed member retrieval (session user not in conversation)",
			StatusCode: http.StatusNotFound,
			Conversation: &models.Conversation{
				ID:          11,
				Name:        "testname",
				Description: utils.StringPtr("testdesc"),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			var userID int64 = 1337
			var memberID int64 = 1
			var conversationID int64 = 11
			if test.SessionMember != nil {
				userID = test.SessionMember.UserID
			}
			if test.ResBody != nil {
				memberID = test.ResBody.UserID
			}

			r := httptest.NewRequest("GET", "/ether/v1/conversations/11/users/1", nil)
			r.Header.Set("User-ID", strconv.FormatInt(userID, 10))
			r = mux.SetURLVars(r, map[string]string{
				"conversation_id": strconv.FormatInt(conversationID, 10),
				"user_id":         strconv.FormatInt(memberID, 10),
			})
			w := httptest.NewRecorder()

			mDB := models.NewMockDB(
				[]*models.Conversation{test.Conversation},
				[]*models.UserConversationMapping{test.SessionMember, test.ResBody},
				nil,
			)

			env := &Env{DB: mDB}
			env.GetMappingHandler(w, r)

			if w.Code != test.StatusCode {
				t.Errorf("Response has incorrect status code, expected status code %d, got %d", test.StatusCode, w.Code)
				return
			}

			if w.Code == http.StatusOK {
				// Validate HTTP response content
				resBody := models.UserConversationMapping{}
				_ = json.NewDecoder(w.Body).Decode(&resBody)
				if !reflect.DeepEqual(*test.ResBody, resBody) {
					t.Errorf("Response has incorrect body, expected %+v, got %+v", *test.ResBody, resBody)
				}
			}
		})
	}
}

func TestGetMappingsHandler(t *testing.T) {
	tests := []struct {
		Name          string
		StatusCode    int
		ResList       []*models.UserConversationMapping
		Conversation  *models.Conversation
		SessionMember *models.UserConversationMapping
	}{
		{
			Name:       "Successful members retrieval (one member)",
			StatusCode: http.StatusOK,
			ResList: []*models.UserConversationMapping{
				&models.UserConversationMapping{
					UserID:         1337,
					ConversationID: 11,
					Role:           "owner",
					Nickname:       utils.StringPtr("testowner"),
					Pending:        utils.BoolPtr(true),
				},
			},
			Conversation: &models.Conversation{
				ID:          11,
				Name:        "testname",
				Description: utils.StringPtr("testdesc"),
			},
			SessionMember: &models.UserConversationMapping{
				UserID:         1337,
				ConversationID: 11,
				Role:           "owner",
				Nickname:       utils.StringPtr("testowner"),
				Pending:        utils.BoolPtr(false),
			},
		},
		{
			Name:       "Successful members retrieval (mulitple members)",
			StatusCode: http.StatusOK,
			ResList: []*models.UserConversationMapping{
				&models.UserConversationMapping{
					UserID:         1337,
					ConversationID: 11,
					Role:           "owner",
					Nickname:       utils.StringPtr("testowner"),
					Pending:        utils.BoolPtr(true),
				},
				&models.UserConversationMapping{
					UserID:         1338,
					ConversationID: 11,
					Role:           "admin",
					Nickname:       utils.StringPtr("testadmin"),
					Pending:        utils.BoolPtr(true),
				},
				&models.UserConversationMapping{
					UserID:         1339,
					ConversationID: 11,
					Role:           "user",
					Nickname:       utils.StringPtr("testuser"),
					Pending:        utils.BoolPtr(true),
				},
			},
			Conversation: &models.Conversation{
				ID:          11,
				Name:        "testname",
				Description: utils.StringPtr("testdesc"),
			},
			SessionMember: &models.UserConversationMapping{
				UserID:         1337,
				ConversationID: 11,
				Role:           "owner",
				Nickname:       utils.StringPtr("testowner"),
				Pending:        utils.BoolPtr(false),
			},
		},
		{
			Name:       "Failed members retrieval (conversation does not exist)",
			StatusCode: http.StatusNotFound,
		},
		{
			Name:       "Failed members retrieval (session user not in conversation)",
			StatusCode: http.StatusNotFound,
			Conversation: &models.Conversation{
				ID:          11,
				Name:        "testname",
				Description: utils.StringPtr("testdesc"),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			var userID int64 = 1337
			var conversationID int64 = 11
			if test.SessionMember != nil {
				userID = test.SessionMember.UserID
			}

			r := httptest.NewRequest("GET", "/ether/v1/conversations/11/users", nil)
			r.Header.Set("User-ID", strconv.FormatInt(userID, 10))
			r = mux.SetURLVars(r, map[string]string{
				"conversation_id": strconv.FormatInt(conversationID, 10),
			})
			w := httptest.NewRecorder()

			mDB := models.NewMockDB(
				[]*models.Conversation{test.Conversation},
				test.ResList,
				nil,
			)

			env := &Env{DB: mDB}
			env.GetMappingsHandler(w, r)

			if w.Code != test.StatusCode {
				t.Errorf("Response has incorrect status code, expected status code %d, got %d", test.StatusCode, w.Code)
				return
			}

			if w.Code == http.StatusOK {
				// Validate HTTP response content
				expectedResBody := models.UserConversationMappingList{
					Users: test.ResList,
				}
				resBody := models.UserConversationMappingList{}
				_ = json.NewDecoder(w.Body).Decode(&resBody)
				if !reflect.DeepEqual(expectedResBody, resBody) {
					t.Errorf("Response has incorrect body, expected %+v, got %+v", expectedResBody, resBody)
				}
			}
		})
	}
}

func TestPatchMappingsHandler(t *testing.T) {
	tests := []struct {
		Name          string
		StatusCode    int
		ReqBody       interface{}
		ResBody       *models.UserConversationMapping
		Conversation  *models.Conversation
		SessionMember *models.UserConversationMapping
		InitialMember *models.UserConversationMapping
	}{
		{
			Name:       "Successful member modification (other member)",
			StatusCode: http.StatusOK,
			ReqBody: map[string]interface{}{
				"role":     "admin",
				"nickname": "testuserMOD",
			},
			ResBody: &models.UserConversationMapping{
				UserID:         1,
				ConversationID: 11,
				Role:           "admin",
				Nickname:       utils.StringPtr("testuserMOD"),
				Pending:        utils.BoolPtr(false),
			},
			Conversation: &models.Conversation{
				ID:          11,
				Name:        "testname",
				Description: utils.StringPtr("testdesc"),
			},
			SessionMember: &models.UserConversationMapping{
				UserID:         1337,
				ConversationID: 11,
				Role:           "owner",
				Nickname:       utils.StringPtr("testowner"),
				Pending:        utils.BoolPtr(false),
			},
			InitialMember: &models.UserConversationMapping{
				UserID:         1,
				ConversationID: 11,
				Role:           "user",
				Nickname:       utils.StringPtr("testuser"),
				Pending:        utils.BoolPtr(false),
			},
		},
		{
			Name:       "Successful member modification (modify self pending)",
			StatusCode: http.StatusOK,
			ReqBody: map[string]interface{}{
				"pending": false,
			},
			ResBody: &models.UserConversationMapping{
				UserID:         1,
				ConversationID: 11,
				Role:           "user",
				Nickname:       utils.StringPtr("testuser"),
				Pending:        utils.BoolPtr(false),
			},
			Conversation: &models.Conversation{
				ID:          11,
				Name:        "testname",
				Description: utils.StringPtr("testdesc"),
			},
			SessionMember: &models.UserConversationMapping{
				UserID:         1,
				ConversationID: 11,
				Role:           "user",
				Nickname:       utils.StringPtr("testuser"),
				Pending:        utils.BoolPtr(true),
			},
		},
		{
			Name:       "Failed member retrieval (member does not exist)",
			StatusCode: http.StatusNotFound,
			Conversation: &models.Conversation{
				ID:          11,
				Name:        "testname",
				Description: utils.StringPtr("testdesc"),
			},
			SessionMember: &models.UserConversationMapping{
				UserID:         1337,
				ConversationID: 11,
				Role:           "owner",
				Nickname:       utils.StringPtr("testowner"),
				Pending:        utils.BoolPtr(false),
			},
		},
		{
			Name:       "Failed members modification (conversation does not exist)",
			StatusCode: http.StatusNotFound,
		},
		{
			Name:       "Failed members modification (session user not in conversation)",
			StatusCode: http.StatusNotFound,
			Conversation: &models.Conversation{
				ID:          11,
				Name:        "testname",
				Description: utils.StringPtr("testdesc"),
			},
		},
		{
			Name:       "Failed member modification (modify pending from false to true)",
			StatusCode: http.StatusForbidden,
			ReqBody: map[string]interface{}{
				"pending": true,
			},
			Conversation: &models.Conversation{
				ID:          11,
				Name:        "testname",
				Description: utils.StringPtr("testdesc"),
			},
			SessionMember: &models.UserConversationMapping{
				UserID:         1,
				ConversationID: 11,
				Role:           "user",
				Nickname:       utils.StringPtr("testuser"),
				Pending:        utils.BoolPtr(false),
			},
		},
		{
			Name:       "Failed member modification (modify other member's pending state)",
			StatusCode: http.StatusForbidden,
			ReqBody: map[string]interface{}{
				"pending": false,
			},
			Conversation: &models.Conversation{
				ID:          11,
				Name:        "testname",
				Description: utils.StringPtr("testdesc"),
			},
			SessionMember: &models.UserConversationMapping{
				UserID:         1337,
				ConversationID: 11,
				Role:           "owner",
				Nickname:       utils.StringPtr("testowner"),
				Pending:        utils.BoolPtr(false),
			},
			InitialMember: &models.UserConversationMapping{
				UserID:         1,
				ConversationID: 11,
				Role:           "user",
				Nickname:       utils.StringPtr("testuser"),
				Pending:        utils.BoolPtr(true),
			},
		},
		{
			Name:       "Failed member modification (session user is pending)",
			StatusCode: http.StatusForbidden,
			ReqBody: map[string]interface{}{
				"nickname": "testuserMOD",
			},
			Conversation: &models.Conversation{
				ID:          11,
				Name:        "testname",
				Description: utils.StringPtr("testdesc"),
			},
			SessionMember: &models.UserConversationMapping{
				UserID:         1337,
				ConversationID: 11,
				Role:           "user",
				Nickname:       utils.StringPtr("testuser"),
				Pending:        utils.BoolPtr(true),
			},
			InitialMember: &models.UserConversationMapping{
				UserID:         1,
				ConversationID: 11,
				Role:           "user",
				Nickname:       utils.StringPtr("testuser"),
				Pending:        utils.BoolPtr(false),
			},
		},
		{
			Name:       "Failed member modification (owner modifies other member's role to owner)",
			StatusCode: http.StatusBadRequest,
			ReqBody: map[string]interface{}{
				"role": "owner",
			},
			Conversation: &models.Conversation{
				ID:          11,
				Name:        "testname",
				Description: utils.StringPtr("testdesc"),
			},
			SessionMember: &models.UserConversationMapping{
				UserID:         1337,
				ConversationID: 11,
				Role:           "owner",
				Nickname:       utils.StringPtr("testowner"),
				Pending:        utils.BoolPtr(false),
			},
			InitialMember: &models.UserConversationMapping{
				UserID:         1,
				ConversationID: 11,
				Role:           "user",
				Nickname:       utils.StringPtr("testuser"),
				Pending:        utils.BoolPtr(false),
			},
		},
		{
			Name:       "Failed member modification (owner modifies own role)",
			StatusCode: http.StatusForbidden,
			ReqBody: map[string]interface{}{
				"role": "user",
			},
			Conversation: &models.Conversation{
				ID:          11,
				Name:        "testname",
				Description: utils.StringPtr("testdesc"),
			},
			SessionMember: &models.UserConversationMapping{
				UserID:         1,
				ConversationID: 11,
				Role:           "owner",
				Nickname:       utils.StringPtr("testowner"),
				Pending:        utils.BoolPtr(false),
			},
		},
		{
			Name:       "Failed member modification (invalid role value)",
			StatusCode: http.StatusBadRequest,
			ReqBody: map[string]interface{}{
				"role": "foobar",
			},
			Conversation: &models.Conversation{
				ID:          11,
				Name:        "testname",
				Description: utils.StringPtr("testdesc"),
			},
			SessionMember: &models.UserConversationMapping{
				UserID:         1337,
				ConversationID: 11,
				Role:           "owner",
				Nickname:       utils.StringPtr("testowner"),
				Pending:        utils.BoolPtr(false),
			},
			InitialMember: &models.UserConversationMapping{
				UserID:         1,
				ConversationID: 11,
				Role:           "user",
				Nickname:       utils.StringPtr("testuser"),
				Pending:        utils.BoolPtr(false),
			},
		},
		{
			Name:       "Failed member modification (non-owner modifies role)",
			StatusCode: http.StatusForbidden,
			ReqBody: map[string]interface{}{
				"role": "admin",
			},
			Conversation: &models.Conversation{
				ID:          11,
				Name:        "testname",
				Description: utils.StringPtr("testdesc"),
			},
			SessionMember: &models.UserConversationMapping{
				UserID:         1337,
				ConversationID: 11,
				Role:           "admin",
				Nickname:       utils.StringPtr("testadmin"),
				Pending:        utils.BoolPtr(false),
			},
			InitialMember: &models.UserConversationMapping{
				UserID:         1,
				ConversationID: 11,
				Role:           "user",
				Nickname:       utils.StringPtr("testuser"),
				Pending:        utils.BoolPtr(false),
			},
		},
		{
			Name:       "Failed member modification (empty JSON)",
			StatusCode: http.StatusBadRequest,
			ReqBody:    map[string]interface{}{},
			Conversation: &models.Conversation{
				ID:          11,
				Name:        "testname",
				Description: utils.StringPtr("testdesc"),
			},
			SessionMember: &models.UserConversationMapping{
				UserID:         1337,
				ConversationID: 11,
				Role:           "owner",
				Nickname:       utils.StringPtr("testowner"),
				Pending:        utils.BoolPtr(false),
			},
			InitialMember: &models.UserConversationMapping{
				UserID:         1,
				ConversationID: 11,
				Role:           "user",
				Nickname:       utils.StringPtr("testuser"),
				Pending:        utils.BoolPtr(false),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			var userID int64 = 1337
			var memberID int64 = 1
			var conversationID int64 = 11
			if test.SessionMember != nil {
				userID = test.SessionMember.UserID
			}

			reqBody, _ := json.Marshal(test.ReqBody)
			r := httptest.NewRequest("PATCH", "/ether/v1/conversations/11/users/1", bytes.NewReader(reqBody))
			r.Header.Set("User-ID", strconv.FormatInt(userID, 10))
			r = mux.SetURLVars(r, map[string]string{
				"conversation_id": strconv.FormatInt(conversationID, 10),
				"user_id":         strconv.FormatInt(memberID, 10),
			})
			w := httptest.NewRecorder()

			mDB := models.NewMockDB(
				[]*models.Conversation{test.Conversation},
				[]*models.UserConversationMapping{test.SessionMember, test.InitialMember},
				nil,
			)

			env := &Env{DB: mDB}
			env.PatchMappingHandler(w, r)

			if w.Code != test.StatusCode {
				t.Errorf("Response has incorrect status code, expected status code %d, got %d", test.StatusCode, w.Code)
				return
			}

			if w.Code == http.StatusOK {
				// Validate HTTP response content
				resBody := models.UserConversationMapping{}
				_ = json.NewDecoder(w.Body).Decode(&resBody)
				if !reflect.DeepEqual(*test.ResBody, resBody) {
					t.Errorf("Response has incorrect body, expected %+v, got %+v", *test.ResBody, resBody)
				}

				if !reflect.DeepEqual(*test.ResBody, *mDB.GetMapping(memberID, conversationID)) {
					t.Errorf(
						"Used incorrect mapping patch, expected %+v, got %+v",
						*test.ResBody,
						mDB.GetMapping(memberID, conversationID),
					)
				}
			}
		})
	}
}

func TestDeleteMappingHandler(t *testing.T) {
	tests := []struct {
		Name          string
		StatusCode    int
		Conversation  *models.Conversation
		SessionMember *models.UserConversationMapping
		Member        *models.UserConversationMapping
	}{
		{
			Name:       "Successful member deletion (self))",
			StatusCode: http.StatusNoContent,
			Conversation: &models.Conversation{
				ID:          11,
				Name:        "testname",
				Description: utils.StringPtr("testdesc"),
			},
			SessionMember: &models.UserConversationMapping{
				UserID:         1,
				ConversationID: 11,
				Role:           "user",
				Nickname:       utils.StringPtr("testuser"),
				Pending:        utils.BoolPtr(false),
			},
		},
		{
			Name:       "Successful member deletion (self while pending))",
			StatusCode: http.StatusNoContent,
			Conversation: &models.Conversation{
				ID:          11,
				Name:        "testname",
				Description: utils.StringPtr("testdesc"),
			},
			SessionMember: &models.UserConversationMapping{
				UserID:         1,
				ConversationID: 11,
				Role:           "user",
				Nickname:       utils.StringPtr("testuser"),
				Pending:        utils.BoolPtr(true),
			},
		},
		{
			Name:       "Successful member deletion (other member)",
			StatusCode: http.StatusNoContent,
			Conversation: &models.Conversation{
				ID:          11,
				Name:        "testname",
				Description: utils.StringPtr("testdesc"),
			},
			SessionMember: &models.UserConversationMapping{
				UserID:         1337,
				ConversationID: 11,
				Role:           "owner",
				Nickname:       utils.StringPtr("testowner"),
				Pending:        utils.BoolPtr(false),
			},
			Member: &models.UserConversationMapping{
				UserID:         1,
				ConversationID: 11,
				Role:           "user",
				Nickname:       utils.StringPtr("testuser"),
				Pending:        utils.BoolPtr(false),
			},
		},
		{
			Name:       "Failed member deletion (member does not exist)",
			StatusCode: http.StatusNotFound,
			Conversation: &models.Conversation{
				ID:          11,
				Name:        "testname",
				Description: utils.StringPtr("testdesc"),
			},
			SessionMember: &models.UserConversationMapping{
				UserID:         1337,
				ConversationID: 11,
				Role:           "owner",
				Nickname:       utils.StringPtr("testowner"),
				Pending:        utils.BoolPtr(false),
			},
		},
		{
			Name:       "Failed member deletion (conversation does not exist)",
			StatusCode: http.StatusNotFound,
		},
		{
			Name:       "Failed member deletion (session user not in conversation)",
			StatusCode: http.StatusNotFound,
			Conversation: &models.Conversation{
				ID:          11,
				Name:        "testname",
				Description: utils.StringPtr("testdesc"),
			},
		},
		{
			Name:       "Failed member deletion (user deletes other user)",
			StatusCode: http.StatusForbidden,
			Conversation: &models.Conversation{
				ID:          11,
				Name:        "testname",
				Description: utils.StringPtr("testdesc"),
			},
			SessionMember: &models.UserConversationMapping{
				UserID:         1337,
				ConversationID: 11,
				Role:           "user",
				Nickname:       utils.StringPtr("testuser"),
				Pending:        utils.BoolPtr(false),
			},
			Member: &models.UserConversationMapping{
				UserID:         1,
				ConversationID: 11,
				Role:           "user",
				Nickname:       utils.StringPtr("testuser"),
				Pending:        utils.BoolPtr(false),
			},
		},
		{
			Name:       "Failed member deletion (pending member deletes other member)",
			StatusCode: http.StatusForbidden,
			Conversation: &models.Conversation{
				ID:          11,
				Name:        "testname",
				Description: utils.StringPtr("testdesc"),
			},
			SessionMember: &models.UserConversationMapping{
				UserID:         1337,
				ConversationID: 11,
				Role:           "admin",
				Nickname:       utils.StringPtr("testadmin"),
				Pending:        utils.BoolPtr(true),
			},
			Member: &models.UserConversationMapping{
				UserID:         1,
				ConversationID: 11,
				Role:           "user",
				Nickname:       utils.StringPtr("testuser"),
				Pending:        utils.BoolPtr(false),
			},
		},
		{
			Name:       "Failed member deletion (owner deletes self)",
			StatusCode: http.StatusForbidden,
			Conversation: &models.Conversation{
				ID:          11,
				Name:        "testname",
				Description: utils.StringPtr("testdesc"),
			},
			SessionMember: &models.UserConversationMapping{
				UserID:         1,
				ConversationID: 11,
				Role:           "owner",
				Nickname:       utils.StringPtr("testowner"),
				Pending:        utils.BoolPtr(false),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			var userID int64 = 1337
			var memberID int64 = 1
			var conversationID int64 = 11
			if test.SessionMember != nil {
				userID = test.SessionMember.UserID
			}

			r := httptest.NewRequest("DELETE", "/ether/v1/conversations/11/users/1", nil)
			r.Header.Set("User-ID", strconv.FormatInt(userID, 10))
			r = mux.SetURLVars(r, map[string]string{
				"conversation_id": strconv.FormatInt(conversationID, 10),
				"user_id":         strconv.FormatInt(memberID, 10),
			})
			w := httptest.NewRecorder()

			mDB := models.NewMockDB(
				[]*models.Conversation{test.Conversation},
				[]*models.UserConversationMapping{test.SessionMember, test.Member},
				nil,
			)

			env := &Env{DB: mDB}
			env.DeleteMappingHandler(w, r)

			if w.Code != test.StatusCode {
				t.Errorf("Response has incorrect status code, expected status code %d, got %d", test.StatusCode, w.Code)
				return
			}
		})
	}
}
