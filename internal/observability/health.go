package observability

import (
	"context"
	"time"

	"github.com/nathanfabio/live-event-streaming-api/internal/middleware"
	"github.com/nathanfabio/live-event-streaming-api/internal/redis"
)


type Service struct {
	redisClient *redis.Client
}

func NewService(client *redis.Client) *Service {
	return &Service{
		redisClient: client,
	}
}

func (s *Service) CheckRedis(ctx context.Context) error {
	return s.redisClient.Ping(ctx).Err()
}

type HealthResponse struct {
	Status string `json:"status"`
	Redis string `json:"redis"`
	Time string `json:"time"`
}

type MetricsResponse struct {
	Requests int64 `json:"requests_total"`
	AvgLatencyMs float64 `json:"avg_latency_ms"`
	UpTme string `json:"uptime"`
}

var startTime = time.Now()

func (s *Service) GetHealth(ctx context.Context) HealthResponse {
	status := "healthy"
	redisStatus := "connected"

	if err := s.CheckRedis(ctx); err != nil {
		status = "unhealthy"
		redisStatus = "disconnected"
	}

	return HealthResponse{
		Status: status,
		Redis: redisStatus,
		Time: time.Now().Format(time.RFC3339),
	}
}

func (s *Service) GetMetrics() MetricsResponse {
	count, totalLat := middleware.GetMetrics()
	avgLat := float64(0)
	if count > 0 {
		avgLat = float64(totalLat.Milliseconds()) / float64(count)
	}
	return MetricsResponse{
		Requests: count,
		AvgLatencyMs: avgLat,
		UpTme: time.Since(startTime).String(),
	}
}