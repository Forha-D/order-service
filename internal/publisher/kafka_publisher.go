package publisher

import (
	"context"
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
func (p *KafkaPublisher) PublishToTopic(
	topic string,
	key string,
	payload []byte,
) error {

	err := p.writer.WriteMessages(
		context.Background(),
		kafka.Message{
			Topic: topic,
			Key:   []byte(key),
			Value: payload,
		},
	)

	if err != nil {
		log.Printf("[publisher] kafka publish failed: %v", err)
		return err
	}

	log.Println("[publisher] event published successfully")

	return nil
}
