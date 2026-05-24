package publisher

import (
	"context"
	"encoding/json"
	"log"

	"github.com/segmentio/kafka-go"
)

type KafkaPublisher struct {
	writer *kafka.Writer
}

func NewKafkaPublisher(
	writer *kafka.Writer,
) *KafkaPublisher {

	return &KafkaPublisher{
		writer: writer,
	}
}

// generic reusable publish function

// NExt: one publish method many topics — pass topic as argument or embed in event envelope
func (p *KafkaPublisher) Publish(
	key string,
	event interface{},
	// topic string,
) error {

	body, err := json.Marshal(event)

	if err != nil {
		return err
	}

	err = p.writer.WriteMessages(
		context.Background(),
		kafka.Message{
			Key:   []byte(key),
			Value: body,
			// Topic: topic,
		},
	)

	if err != nil {
		log.Printf("[publisher] kafka publish failed: %v", err)
		return err
	}

	log.Println("[publisher] event published successfully")

	return nil
}
