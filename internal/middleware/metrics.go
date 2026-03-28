package middleware

import (
	"net/http"
	"sync"
	"time"
)

//Metrics holds in-memory metrics for the custom endpoind
//In prod, this would be pushed to Prometheus or similar, but I expose via JSON per request

var (
	metricsMu sync.RWMutex
	reqCount int64
	totalLatency time.Duration
)

func Metrics() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			duration := time.Since(start)

			metricsMu.Lock()
			reqCount++
			totalLatency += duration
			metricsMu.Unlock()
		})
	}
}

func GetMetrics() (int64, time.Duration) {
	metricsMu.RLock()
	defer metricsMu.RUnlock()
	return reqCount, totalLatency
}