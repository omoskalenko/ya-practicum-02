package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/segmentio/kafka-go"
)

type Config struct {
	Port         string
	KafkaBrokers []string
}

var config Config
var kafkaWriters map[string]*kafka.Writer
var kafkaReader *kafka.Reader

// Event models
type MovieEvent struct {
	MovieID     int      `json:"movie_id"`
	Title       string   `json:"title"`
	Action      string   `json:"action"`
	UserID      *int     `json:"user_id,omitempty"`
	Rating      *float64 `json:"rating,omitempty"`
	Genres      []string `json:"genres,omitempty"`
	Description *string  `json:"description,omitempty"`
}

type UserEvent struct {
	UserID    int    `json:"user_id"`
	Username  string `json:"username,omitempty"`
	Email     string `json:"email,omitempty"`
	Action    string `json:"action"`
	Timestamp string `json:"timestamp"`
}

type PaymentEvent struct {
	PaymentID  int     `json:"payment_id"`
	UserID     int     `json:"user_id"`
	Amount     float64 `json:"amount"`
	Status     string  `json:"status"`
	Timestamp  string  `json:"timestamp"`
	MethodType string  `json:"method_type,omitempty"`
}

type EventResponse struct {
	Status    string      `json:"status"`
	Partition int         `json:"partition"`
	Offset    int64       `json:"offset"`
	Event     interface{} `json:"event"`
}

func main() {
	loadConfig()
	setupKafkaProducers()

	// Start Kafka consumer in background for all topics
	go startKafkaConsumerForAllTopics()

	// Setup HTTP routes
	http.HandleFunc("/api/events/health", handleHealth)
	http.HandleFunc("/api/events/movie", handleMovieEvent)
	http.HandleFunc("/api/events/user", handleUserEvent)
	http.HandleFunc("/api/events/payment", handlePaymentEvent)

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	server := &http.Server{
		Addr: ":" + config.Port,
	}

	go func() {
		log.Printf("Starting Events Service on port %s", config.Port)
		log.Printf("Kafka brokers: %v", config.KafkaBrokers)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-sigChan
	log.Println("Shutting down gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	closeKafkaProducers()
	if kafkaReader != nil {
		kafkaReader.Close()
	}

	server.Shutdown(ctx)
}

func loadConfig() {
	config.Port = getEnv("PORT", "8082")

	brokersStr := getEnv("KAFKA_BROKERS", "kafka:9092")
	config.KafkaBrokers = strings.Split(brokersStr, ",")
}

func setupKafkaProducers() {
	kafkaWriters = make(map[string]*kafka.Writer)

	topics := []string{"movie-events", "user-events", "payment-events"}

	for _, topic := range topics {
		writer := &kafka.Writer{
			Addr:         kafka.TCP(config.KafkaBrokers...),
			Topic:        topic,
			Balancer:     &kafka.LeastBytes{},
			BatchTimeout: 10 * time.Millisecond,
			RequiredAcks: kafka.RequireOne,
		}
		kafkaWriters[topic] = writer
		log.Printf("✓ Kafka producer initialized for topic: %s", topic)
	}
}

func closeKafkaProducers() {
	for topic, writer := range kafkaWriters {
		if err := writer.Close(); err != nil {
			log.Printf("Error closing Kafka writer for %s: %v", topic, err)
		}
	}
}

func startKafkaConsumerForAllTopics() {
	// Consumer Group API - 3 consumer'а в одной группе читают из разных топиков
	// Это создаст все топики автоматически при подключении
	topics := []string{"movie-events", "user-events", "payment-events"}

	log.Printf("Starting Kafka Consumer Group: events-service-consumer")
	log.Printf("Subscribing to topics: %v", topics)

	for _, topic := range topics {
		go consumeTopic(topic)
	}
}

func consumeTopic(topic string) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  config.KafkaBrokers,
		GroupID:  "events-service-consumer", // Все в одной группе
		Topic:    topic,
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})

	log.Printf("✓ Consumer started for topic: %s (group: events-service-consumer)", topic)

	for {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("❌ Error reading from %s: %v", topic, err)
			time.Sleep(1 * time.Second)
			continue
		}

		log.Printf("✓ [CONSUMED] Topic: %s | Partition: %d | Offset: %d | Key: %s | Value: %s",
			msg.Topic, msg.Partition, msg.Offset, string(msg.Key), string(msg.Value))
	}
}

func publishEvent(topic string, key string, event interface{}) (*EventResponse, error) {
	writer, exists := kafkaWriters[topic]
	if !exists {
		return nil, fmt.Errorf("no writer for topic: %s", topic)
	}

	eventJSON, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}

	msg := kafka.Message{
		Key:   []byte(key),
		Value: eventJSON,
		Time:  time.Now(),
	}

	err = writer.WriteMessages(context.Background(), msg)
	if err != nil {
		return nil, err
	}

	log.Printf("[PUBLISHED] Topic: %s, Key: %s, Event: %s", topic, key, string(eventJSON))

	response := &EventResponse{
		Status:    "success",
		Partition: 0,
		Offset:    0,
		Event:     event,
	}

	return response, nil
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"status": true})
}

func handleMovieEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var event MovieEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	key := fmt.Sprintf("movie-%d", event.MovieID)
	response, err := publishEvent("movie-events", key, event)
	if err != nil {
		log.Printf("Error publishing movie event: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func handleUserEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var event UserEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	key := fmt.Sprintf("user-%d", event.UserID)
	response, err := publishEvent("user-events", key, event)
	if err != nil {
		log.Printf("Error publishing user event: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func handlePaymentEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var event PaymentEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	key := fmt.Sprintf("payment-%d", event.PaymentID)
	response, err := publishEvent("payment-events", key, event)
	if err != nil {
		log.Printf("Error publishing payment event: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
