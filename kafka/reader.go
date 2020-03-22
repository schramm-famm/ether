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

type Reader struct {
	reader       *kafka.Reader
	cachedWriter *filesystem.CachedWriter
	db           models.Datastore
}

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

func (r *Reader) Run() {
	defer r.reader.Close()

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

		update := &filesystem.Update{
			ConversationID: conversationID,
			Patch:          *message.Data.Patch,
		}

		r.cachedWriter.Write <- update

		err = r.db.TouchConversation(conversationID)
		if err != nil {
			log.Printf("Failed to update LastModified time: %v", err)
			continue
		}
	}
}
