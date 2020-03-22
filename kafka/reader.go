package kafka

import (
	"context"
	"encoding/json"
	"ether/filesystem"
	"ether/models"
	"log"
	"strconv"

	segkafka "github.com/segmentio/kafka-go"
)

// Reader represents a Kafka consumer which consumes and processes conversation
// update messages.
type Reader struct {
	reader       *segkafka.Reader
	cachedWriter *filesystem.CachedWriter
	db           models.Datastore
}

// NewReader initializes a new Reader.
func NewReader(location, topic string, directory *filesystem.Directory, db models.Datastore) *Reader {
	return &Reader{
		reader: segkafka.NewReader(segkafka.ReaderConfig{
			Brokers:  []string{location},
			GroupID:  "ether",
			Topic:    topic,
			MinBytes: 1,
			MaxBytes: 10e6,
		}),
		cachedWriter: filesystem.NewCachedWriter(directory),
		db:           db,
	}
}

func (r *Reader) handleUpdate(conversationID int64, msg Message) error {
	// Tell writer goroutine to update this conversation's content file with
	// this patch
	update := &filesystem.Update{
		ConversationID: conversationID,
		Patch:          *msg.Data.Patch,
	}
	r.cachedWriter.Write <- update

	// Set conversation LastModified time to now
	return r.db.TouchConversation(conversationID)
}

func (r *Reader) processMessage(kafkaMsg segkafka.Message) error {
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
		if err := r.handleUpdate(conversationID, msg); err != nil {
			return err
		}

	default:
		// TODO: handle other message types (UserJoin, UserLeave)?
	}

	return nil
}

// Run reads from the Kafka topic indefinitely and, upon receiving a message,
// updates the relevant conversation content file and updates the relevant
// conversation LastModified time in the database.
func (r *Reader) Run() {
	defer r.reader.Close()

	// Start file writer goroutine
	go r.cachedWriter.Run()

	for {
		m, err := r.reader.ReadMessage(context.Background())
		if err != nil {
			log.Fatal(err)
		}

		if err := r.processMessage(m); err != nil {
			log.Printf("Failed to process Kafka message: %v", err)
		}
	}
}
