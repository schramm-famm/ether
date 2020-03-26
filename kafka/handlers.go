package kafka

import (
	"encoding/json"
	"ether/filesystem"
	"ether/models"
	"strconv"

	segkafka "github.com/segmentio/kafka-go"
)

// Env represents all application-level items that are needed by Kafka handlers.
type Env struct {
	DB           models.Datastore
	CachedWriter *filesystem.CachedWriter
}

// processUpdate processes an Update type Kafka message for a given
// conversation.
func (env *Env) processUpdate(conversationID int64, msg Message) error {
	// Tell writer goroutine to update this conversation's content file with
	// this patch
	update := &filesystem.Update{
		ConversationID: conversationID,
		Patch:          *msg.Data.Patch,
	}
	env.CachedWriter.Write <- update

	// Set conversation LastModified time to now
	return env.DB.TouchConversation(conversationID)
}

// ProcessWSMessage processes a Kafka message that corresponds to a WebSocket
// message being handled by the "patches" service.
func (env *Env) ProcessWSMessage(kafkaMsg segkafka.Message) error {
	conversationID, err := strconv.ParseInt(string(kafkaMsg.Key), 10, 64)
	if err != nil {
		return err
	}

	msg := Message{}
	if err := json.Unmarshal(kafkaMsg.Value, &msg); err != nil {
		return err
	}

	switch msg.Type {
	case TypeUpdate:
		if err := env.processUpdate(conversationID, msg); err != nil {
			return err
		}

	default:
		// TODO: handle other message types (UserJoin, UserLeave)?
	}

	return nil
}
