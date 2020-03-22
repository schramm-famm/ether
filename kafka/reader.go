package kafka

import (
	"context"
	"encoding/json"
	"ether/filesystem"
	"ether/models"
	"log"
	"strconv"

	kafka "github.com/segmentio/kafka-go"
)

// Reader represents a Kafka consumer which consumes and processes conversation
// update messages.
type Reader struct {
	reader       *kafka.Reader
	cachedWriter *filesystem.CachedWriter
	db           models.Datastore
}

// NewReader initializes a new Reader.
func NewReader(location, topic string, directory *filesystem.Directory, db models.Datastore) *Reader {
	return &Reader{
		reader: kafka.NewReader(kafka.ReaderConfig{
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

		conversationID, err := strconv.ParseInt(string(m.Key), 10, 64)
		if err != nil {
			log.Printf("Kafka message key was not an integer: %s", string(m.Key))
			continue
		}

		message := Message{}
		if err := json.Unmarshal(m.Value, &message); err != nil {
			log.Printf("Kafka message value malformed: %v", err)
			continue
		}

		// Tell writer goroutine to update this conversation's content file with
		// this patch
		update := &filesystem.Update{
			ConversationID: conversationID,
			Patch:          *message.Data.Patch,
		}
		r.cachedWriter.Write <- update

		// Set conversation LastModified time to now
		err = r.db.TouchConversation(conversationID)
		if err != nil {
			log.Printf("Failed to update LastModified time: %v", err)
			continue
		}
	}
}
