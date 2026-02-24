package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/oragazz0/nexus-event-stream/data-plane/internal/consumer"
	"github.com/oragazz0/nexus-event-stream/data-plane/internal/handler"
	"github.com/oragazz0/nexus-event-stream/data-plane/internal/projection"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

func main() {
	ctx := setupContext()

	redisClient := connectRedis(ctx)
	defer redisClient.Close()

	proj := projection.New(redisClient)

	startConsumer(ctx, proj)
	serveHTTP(ctx, proj)
}

func setupContext() context.Context {
	ctx, _ := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	return ctx
}

func connectRedis(ctx context.Context) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr: envOrDefault("REDIS_ADDR", "localhost:6379"),
	})
	pingResult := client.Ping(ctx)
	if err := pingResult.Err(); err != nil {
		log.Fatalf("redis connection failed: %v", err)
	}
	log.Println("connected to redis")
	return client
}

func startConsumer(ctx context.Context, proj projection.SignalProjection) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{envOrDefault("KAFKA_BROKERS", "localhost:9092")},
		Topic:       "nexus.signals",
		GroupID:     "nexus-data-plane",
		StartOffset: kafka.FirstOffset,
	})
	cons := consumer.New(reader, proj)
	go func() {
		log.Println("consumer started")
		defer reader.Close()
		if err := cons.Start(ctx); err != nil {
			log.Printf("consumer stopped: %v", err)
		}
	}()
}

func serveHTTP(ctx context.Context, proj projection.SignalProjection) {
	signalHandler := handler.New(proj)
	mux := http.NewServeMux()
	signalHandler.Register(mux)

	addr := envOrDefault("HTTP_ADDR", ":8081")
	server := &http.Server{Addr: addr, Handler: mux}

	go func() {
		<-ctx.Done()
		server.Shutdown(context.Background())
	}()

	log.Printf("http server listening on %s", addr)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("http server error: %v", err)
	}
	log.Println("shutdown complete")
}

func envOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value != "" {
		return value
	}
	return fallback
}
