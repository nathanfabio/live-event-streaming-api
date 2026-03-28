package semaphore

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/nathanfabio/live-event-streaming-api/internal/redis"
)


var ErrConcurrencyLimitExceeded = errors.New("user concurrency limit exceeded")

type RedisSemaphore struct {
	client *redis.Client
}


func NewRedisSemaphore(client *redis.Client) *RedisSemaphore {
	return &RedisSemaphore{client: client}
}


//Acquire attempts to acquire a semaphore for the given user ID. It returns an error if the concurrency limit is exceeded.
func (s *RedisSemaphore) Acquire(ctx context.Context, userID string, limit int) error {
	key := fmt.Sprintf("stream:concurrency:%s", userID)

	//Use Redis INCR command to atomically increment the count of active streams for the user
	val, err := s.client.Incr(ctx, key).Result()
	if err != nil {
		return err
	}

	if val > int64(limit) {
		//If the count exceeds the limit, decrement it back and return an error
		s.client.Decr(ctx, key)
		return ErrConcurrencyLimitExceeded
	}

	//Set an expiration for the key to automatically release the semaphore after a certain time (e.g., 1 hour)
	s.client.Expire(ctx, key, 1*time.Hour)
	return nil

}

//Release releases the semaphore for the given user ID by decrementing the count of active streams.
func (s *RedisSemaphore) Release(ctx context.Context, userID string) error {
	key := fmt.Sprintf("stream:concurrency:%s", userID)
	_, err := s.client.Decr(ctx, key).Result()
	return err
}