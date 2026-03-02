package main

import (
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type ProxyConfig struct {
	Port                   string
	MonolithURL            *url.URL
	MoviesServiceURL       *url.URL
	EventsServiceURL       *url.URL
	GradualMigration       bool
	MoviesMigrationPercent int
}

var config ProxyConfig

func main() {
	rand.Seed(time.Now().UnixNano())

	if err := loadConfig(); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Starting Proxy Service (API Gateway) on port %s", config.Port)
	log.Printf("Monolith URL: %s", config.MonolithURL)
	log.Printf("Movies Service URL: %s", config.MoviesServiceURL)
	log.Printf("Events Service URL: %s", config.EventsServiceURL)
	log.Printf("Gradual Migration: %v", config.GradualMigration)
	log.Printf("Movies Migration Percent: %d%%", config.MoviesMigrationPercent)

	http.HandleFunc("/", handleRequest)

	log.Fatal(http.ListenAndServe(":"+config.Port, nil))
}

func loadConfig() error {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	config.Port = port

	monolithURL, err := url.Parse(getEnv("MONOLITH_URL", "http://monolith:8080"))
	if err != nil {
		return err
	}
	config.MonolithURL = monolithURL

	moviesServiceURL, err := url.Parse(getEnv("MOVIES_SERVICE_URL", "http://movies-service:8081"))
	if err != nil {
		return err
	}
	config.MoviesServiceURL = moviesServiceURL

	eventsServiceURL, err := url.Parse(getEnv("EVENTS_SERVICE_URL", "http://events-service:8082"))
	if err != nil {
		return err
	}
	config.EventsServiceURL = eventsServiceURL

	config.GradualMigration = getEnv("GRADUAL_MIGRATION", "true") == "true"

	migrationPercent, err := strconv.Atoi(getEnv("MOVIES_MIGRATION_PERCENT", "50"))
	if err != nil {
		migrationPercent = 50
	}
	config.MoviesMigrationPercent = migrationPercent

	return nil
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Determine target URL based on path and feature flags
	var targetURL *url.URL
	var serviceName string

	if strings.HasPrefix(path, "/api/events") {
		// Route events to events service
		targetURL = config.EventsServiceURL
		serviceName = "events-service"
	} else if strings.HasPrefix(path, "/api/movies") {
		// Strangler Fig pattern: gradually migrate movies traffic
		if config.GradualMigration && shouldRouteToMicroservice() {
			targetURL = config.MoviesServiceURL
			serviceName = "movies-service"
		} else {
			targetURL = config.MonolithURL
			serviceName = "monolith"
		}
	} else {
		// Route everything else to monolith
		targetURL = config.MonolithURL
		serviceName = "monolith"
	}

	log.Printf("[%s] %s %s -> %s", serviceName, r.Method, path, targetURL)

	// Create reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// Modify the request to include the original host
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = targetURL.Host
		req.URL.Scheme = targetURL.Scheme
		req.URL.Host = targetURL.Host
	}

	// Handle errors
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("Proxy error for %s: %v", path, err)
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
	}

	// Proxy the request
	proxy.ServeHTTP(w, r)
}

// shouldRouteToMicroservice determines if request should go to microservice
// based on migration percentage (Strangler Fig pattern)
func shouldRouteToMicroservice() bool {
	randomPercent := rand.Intn(100)
	return randomPercent < config.MoviesMigrationPercent
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
