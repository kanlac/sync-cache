package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/kanlac/sync-cache/cache"
	"github.com/kanlac/sync-cache/engine"
)

type Server struct {
	cache cache.SyncCache[string, string]
}

type SetRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type GetResponse struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Found bool   `json:"found"`
}

type HealthResponse struct {
	Status string `json:"status"`
	Time   string `json:"time"`
}

func main() {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	instanceName := os.Getenv("INSTANCE_NAME")
	if instanceName == "" {
		instanceName = "default-instance"
	}

	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8080"
	}

	// Create sync cache instance
	syncCache, err := engine.NewRistrettoCacheEngine[string, string](redisAddr, instanceName)
	if err != nil {
		log.Fatalf("Failed to create sync cache: %v", err)
	}

	server := &Server{cache: syncCache}

	// Setup HTTP routes
	http.HandleFunc("/health", server.healthHandler)
	http.HandleFunc("/set", server.setHandler)
	http.HandleFunc("/get", server.getHandler)
	http.HandleFunc("/delete", server.deleteHandler)

	log.Printf("Starting server on port %s with instance name: %s", httpPort, instanceName)
	log.Fatal(http.ListenAndServe(":"+httpPort, nil))
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := HealthResponse{
		Status: "ok",
		Time:   time.Now().UTC().Format(time.RFC3339),
	}
	json.NewEncoder(w).Encode(response)
}

func (s *Server) setHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Key == "" {
		http.Error(w, "Key is required", http.StatusBadRequest)
		return
	}

	s.cache.Set(req.Key, req.Value)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) getHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "Key parameter is required", http.StatusBadRequest)
		return
	}

	value, found := s.cache.Get(key)

	w.Header().Set("Content-Type", "application/json")
	response := GetResponse{
		Key:   key,
		Value: value,
		Found: found,
	}
	json.NewEncoder(w).Encode(response)
}

func (s *Server) deleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "Key parameter is required", http.StatusBadRequest)
		return
	}

	s.cache.Delete(key)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
