package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/nathanfabio/live-event-streaming-api/internal/config"
	"github.com/nathanfabio/live-event-streaming-api/internal/handler"
	"github.com/nathanfabio/live-event-streaming-api/internal/middleware"
	"github.com/nathanfabio/live-event-streaming-api/internal/observability"
	"github.com/nathanfabio/live-event-streaming-api/internal/redis"
	"github.com/nathanfabio/live-event-streaming-api/internal/service"
	"github.com/nathanfabio/live-event-streaming-api/pkg/logger"
	"go.uber.org/zap"
)


func main() {
	zapLogger, err := logger.New()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer zapLogger.Sync()

	cfg, err := config.Load()
	if err != nil {
		zapLogger.Fatal("Failed to load configuration", zap.Error(err))
	}

	redisClient, err := redis.NewClient(cfg.RedisURL)
	if err != nil {
		zapLogger.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	defer redisClient.Close()

	streamService := service.NewStreamService(redisClient, cfg)
	obsService := observability.NewService(redisClient)


	streamHandler := handler.NewStreamHandler(streamService, zapLogger)
	healthHandler := handler.NewHealthyHandler(obsService, zapLogger)

	r := mux.NewRouter()

	r.HandleFunc("/health", healthHandler.Health).Methods("GET")
	r.HandleFunc("/ready", healthHandler.Ready).Methods("GET")
	r.HandleFunc("/metrics", healthHandler.Metrics).Methods("GET")

	api := r.PathPrefix("/api/v1").Subrouter()
	//api.Use(middleware.Logging(zapLogger))
	api.Use(middleware.Auth(cfg.JWTSecret))
	api.Use(middleware.Metrics())

	api.HandleFunc("/stream/{id}/play", streamHandler.PlayStream).Methods("GET")
	api.HandleFunc("/stream/{id}/manifest", streamHandler.GetManifest).Methods("GET")


	srv := &http.Server{
		Addr: fmt.Sprintf(":%s", cfg.Port),
		Handler: r,
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout: 60 * time.Second,
	}

	go func () {
		zapLogger.Info("Starting server", zap.String("port", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zapLogger.Fatal("Server failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	zapLogger.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		zapLogger.Fatal("Server forced to shutdown", zap.Error(err))
	}
	zapLogger.Info("Server exiting")

}