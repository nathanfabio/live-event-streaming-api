package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/nathanfabio/live-event-streaming-api/internal/config"
	"github.com/nathanfabio/live-event-streaming-api/internal/pkg/semaphore"
	"github.com/nathanfabio/live-event-streaming-api/internal/redis"
)


type StreamService struct {
	redisClient *redis.Client
	semaphore *semaphore.RedisSemaphore
	config *config.Config
}

type StreamMetadata struct {
	PrimaryCDN   string
	SecondaryCDN string
}


func NewStreamService(client *redis.Client, cfg *config.Config) *StreamService {
	return &StreamService{
		redisClient: client,
		semaphore: semaphore.NewRedisSemaphore(client),
		config: cfg,
	}
}


//PlayStream handles the logic to authorize and start a stream session
func (s *StreamService) PlayStream(ctx context.Context, userID, streamID string) (string, error) {
	if err := s.semaphore.Acquire(ctx, userID, s.config.MaxConcurrentStreamsPerUser); err != nil {
		if errors.Is(err, semaphore.ErrConcurrencyLimitExceeded) {
			return "", fmt.Errorf("limit reached: %w", err)
		}
		return "", err
	}

	metadata, err := s.getStreamMetaData(ctx, streamID)
	if err != nil {
		s.semaphore.Release(ctx, userID)
		return "", err
	}

	streamURL, err := s.resolveStreamURL(ctx, metadata.PrimaryCDN, metadata.SecondaryCDN)
	if err != nil {
		s.semaphore.Release(ctx, userID)
		return "", err
	}
	return streamURL, nil
}

//getStreamMetadata demonstrates caching pattern
func (s *StreamService) getStreamMetaData(ctx context.Context, streamID string) (*StreamMetadata, error) {
	cacheKey := fmt.Sprintf("stream:metadata:%s", streamID)

	_, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		return &StreamMetadata{PrimaryCDN: s.config.PrimaryProviderURL, SecondaryCDN: s.config.SecondaryProviderURL}, nil
	}

	// Cache Miss - Mock DB fetch
	meta := &StreamMetadata{PrimaryCDN: s.config.PrimaryProviderURL, SecondaryCDN: s.config.SecondaryProviderURL}
	
	// Set Cache (10 minutes)
	s.redisClient.Set(ctx, cacheKey, "mock_data", 10*time.Minute)
	
	return meta, nil
}


// resolveStreamURL demonstrates Fallback Logic
func (s *StreamService) resolveStreamURL(ctx context.Context, primary, secondary string) (string, error) {
	// Simulate Health Check on Primary
	// In real app, this would be an HTTP HEAD request or circuit breaker check
	isPrimaryHealthy := true // Mocked
	
	if isPrimaryHealthy {
		return primary, nil
	}

	// Fallback to Secondary
	if secondary != "" {
		return secondary, nil
	}

	return "", errors.New("no available stream providers")
}


