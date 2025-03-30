package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"

	localRedis "github.com/yourusername/redis-demo/backend/db/redis"
	"github.com/yourusername/redis-demo/backend/redistest"
)

var (
	redisClient *redis.Client
)

func setupRouter(redisClient *redis.Client) *gin.Engine {
	r := gin.Default()

	// Create a redistest.Handler with the Redis client
	handler := redistest.NewHandler(redisClient)

	// Register the routes
	r.GET("/txpipeline", handler.TxPipelineDemo)
	r.GET("/syntax-error", handler.SyntaxErrorDemo)
	r.GET("/logic-error", handler.LogicErrorDemo)

	return r
}

func cleanupResources() {
	if redisClient != nil {
		log.Println("Closing Redis connection...")
		if err := redisClient.Close(); err != nil {
			log.Printf("Error closing Redis connection: %v", err)
		}
	}
	log.Println("Resources cleaned up")
}

func main() {
	var err error
	redisClient, err = localRedis.NewClient(localRedis.DefaultConfig())
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Connected to Redis successfully")
	defer cleanupResources()

	// Setup router
	r := setupRouter(redisClient)

	// Create a server with graceful shutdown
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// Start the server in a goroutine
	go func() {
		log.Println("Server starting on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Create a deadline for server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}
