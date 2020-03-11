package handlers

import (
	"ether/models"
	"ether/utils"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strconv"
	"testing"

	"github.com/gorilla/mux"
)

func TestGetContentHandler(t *testing.T) {
	tests := []struct {
		Name         string
		StatusCode   int
		Content      string
		Conversation *models.Conversation
		Mapping      *models.UserConversationMapping
	}{
		{
			Name:       "Successful content retrieval",
			StatusCode: http.StatusOK,
			Content:    "hello world",
			Conversation: &models.Conversation{
				ID:          1,
				Name:        "test_name",
				Description: utils.StringPtr("test_desc"),
				AvatarURL:   utils.StringPtr("test_url"),
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
			Name:       "Failed content retrieval (conversation does not exist)",
			StatusCode: http.StatusNotFound,
		},
		{
			Name:       "Failed content retrieval (user not in conversation)",
			StatusCode: http.StatusNotFound,
			Conversation: &models.Conversation{
				ID:          1,
				Name:        "test_name",
				Description: utils.StringPtr("test_desc"),
				AvatarURL:   utils.StringPtr("test_url"),
			},
		},
		{
			Name:       "Failed content retrieval (pending invitation)",
			StatusCode: http.StatusForbidden,
			Conversation: &models.Conversation{
				ID:          1,
				Name:        "test_name",
				Description: utils.StringPtr("test_desc"),
				AvatarURL:   utils.StringPtr("test_url"),
			},
			Mapping: &models.UserConversationMapping{
				UserID:         1,
				ConversationID: 1,
				Role:           "user",
				Nickname:       utils.StringPtr(""),
				Pending:        utils.BoolPtr(true),
				LastOpened:     "2006-01-02 15:04:05",
			},
		},
	}

	var userID int64 = 1
	var conversationID int64 = 1
	filePath := path.Join(contentDir, fmt.Sprintf("%d.html", conversationID))
	f, _ := os.Create(filePath)
	f.WriteString("hello world")
	f.Close()
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/ether/v1/conversations/1/content", nil)
			r.Header.Set("User-ID", strconv.FormatInt(userID, 10))
			r = mux.SetURLVars(r, map[string]string{
				"conversation_id": strconv.FormatInt(conversationID, 10),
			})
			w := httptest.NewRecorder()

			mDB := models.NewMockDB(
				[]*models.Conversation{test.Conversation},
				[]*models.UserConversationMapping{test.Mapping},
				nil,
			)

			env := &Env{DB: mDB}
			env.GetContentHandler(w, r)

			if w.Code != test.StatusCode {
				t.Errorf("Response has incorrect status code, expected status code %d, got %d", test.StatusCode, w.Code)
			}

			if w.Code == http.StatusOK {
				// Validate HTTP response content
				resBody, _ := ioutil.ReadAll(w.Body)
				resContent := string(resBody)
				if test.Content != resContent {
					t.Errorf("Response has incorrect body, expected %q, got %q", test.Content, resContent)
				}
			}
		})
	}
	os.Remove(filePath)
}
