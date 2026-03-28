package handler

import (
	"encoding/json"
	"net/http"

	"github.com/nathanfabio/live-event-streaming-api/internal/observability"
	"go.uber.org/zap"
)


type HealthyHandler struct {
	service *observability.Service
	logger *zap.Logger
}

func NewHealthyHandler(s *observability.Service, l *zap.Logger) *HealthyHandler {
	return &HealthyHandler{
		service: s,
		logger: l,
	}
}


func (h *HealthyHandler) Health(w http.ResponseWriter, r *http.Request) {
	health := h.service.GetHealth(r.Context())
	statusCode := http.StatusOK
	if health.Status != "healthy" {
		statusCode = http.StatusInternalServerError
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(health)
}

func (h *HealthyHandler) Ready(w http.ResponseWriter, r *http.Request) {
	//Similar to health but stricter checks can be added here (db migrations, etc.)
	h.Health(w, r)
}

func (h *HealthyHandler) Metrics(w http.ResponseWriter, r *http.Request) {
	metrics := h.service.GetMetrics()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}
