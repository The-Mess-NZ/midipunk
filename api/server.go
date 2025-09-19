package api

import (
	"log"
	"net/http"

	"github.com/The-Mess-NZ/midipunk/api/handlers"
	"github.com/The-Mess-NZ/midipunk/logger"
)

func StartAPIServer() {
	http.HandleFunc("/config/get", handlers.ConfigGetHandler)
	http.HandleFunc("/config/put", handlers.ConfigPutHandler)
	http.HandleFunc("/status/ws", handlers.StatusWSHandler)
	http.HandleFunc("/events/ws", handlers.EventWSHandler)

	logger.Info("API server starting on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		logger.Error("API server failed: %v", err)
		log.Fatalf("API server failed: %v", err)
	}
}
