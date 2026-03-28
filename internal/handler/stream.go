package handler

import (
	"net/http"

	"github.com/nathanfabio/live-event-streaming-api/internal/service"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)


type StreamHandler struct {
	service *service.StreamService
	logger *zap.Logger
}

func NewStreamHandler(s *service.StreamService, l *zap.Logger) *StreamHandler {
	return &StreamHandler{
		service: s,
		logger: l,
	}
}

func (h *StreamHandler) PlayStream(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	streamID := vars["id"]

	userID := r.Context().Value("userID").(string)

	url, err := h.service.PlayStream(r.Context(), userID, streamID)
	if err != nil {
		h.logger.Warn("Stream play failed", zap.Error(err), zap.String("user", userID))
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	w.Write([]byte(url))
}

func (h *StreamHandler) GetManifest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	w.Header().Set("Cache-Control", "max-age=2, public")
	w.Write([]byte("#EXTM3U\n#EXT-X-VERSION:3\n#EXT-X-TARGETDURATION:4\n#EXTINF:4.0,\nsegment.ts\n"))
}