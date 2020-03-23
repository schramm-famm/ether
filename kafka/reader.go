package kafka

import (
	"context"
	"log"

	segkafka "github.com/segmentio/kafka-go"
)

// Reader represents a Kafka consumer which consumes and processes conversation
// update messages.
type Reader struct {
	reader *segkafka.Reader
}

// NewReader initializes a new Reader.
func NewReader(location, topic string) *Reader {
	return &Reader{
		reader: segkafka.NewReader(segkafka.ReaderConfig{
			Brokers:  []string{location},
			GroupID:  "ether",
			Topic:    topic,
			MinBytes: 1,
			MaxBytes: 10e6,
		}),
	}
}

// Run reads from the Kafka topic indefinitely and, upon receiving a message,
// updates the relevant conversation content file and updates the relevant
// conversation LastModified time in the database.
func (r *Reader) Run(handler func(m segkafka.Message) error) {
	defer r.reader.Close()

	for {
		m, err := r.reader.ReadMessage(context.Background())
		if err != nil {
			log.Fatal(err)
		}

		if err := handler(m); err != nil {
			log.Printf("Failed to process Kafka message: %v", err)
		}
	}
}
