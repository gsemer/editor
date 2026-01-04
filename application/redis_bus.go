package application

import (
	"context"
	"editor/domain"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type MessageBroker interface {
	Publish(documentID string, delta *domain.Delta) error
	Subscribe(documentID string, handler func(domain.Delta))
}

type RedisBus struct {
	rdb *redis.Client
}

func NewRedisBus(rdb *redis.Client) *RedisBus {
	return &RedisBus{rdb: rdb}
}

func (bus *RedisBus) Publish(documentID string, delta *domain.Delta) error {
	message, err := json.Marshal(delta)
	if err != nil {
		log.Printf("Error marshaling: %v", err)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	bus.rdb.Publish(ctx, fmt.Sprintf("document:%s", documentID), message)

	return nil
}

func (bus *RedisBus) Subscribe(documentID string, handler func(domain.Delta)) {
	pubsub := bus.rdb.Subscribe(context.Background(), fmt.Sprintf("document:%s", documentID))

	go func() {
		defer pubsub.Close()
		ch := pubsub.Channel()
		for msg := range ch {
			var d domain.Delta
			if err := json.Unmarshal([]byte(msg.Payload), &d); err == nil {
				// Execute the handler provided by the application layer
				handler(d)
			}
		}
	}()
}
