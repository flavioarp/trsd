package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type ModeRequest struct {
	Mode string `json:"mode"`
}

type KnowledgeState struct {
	Fallback bool `json:"fallback"`
}

// Sends the mode decision to the core service
func setModeOnCore(mode string) error {
	modeReq := ModeRequest{Mode: mode}
	modeBody, _ := json.Marshal(modeReq)
	resp, err := http.Post("http://core:8082/mode", "application/json", bytes.NewBuffer(modeBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Failed to set mode on core: status %d\n", resp.StatusCode)
		return err
	}

	log.Printf("Set core mode to: %s\n", mode)
	return nil
}

// determineModeFromServices queries the knowledge service to determine the mode
// The plan service only forwards writes to the knowledge service, so querying
// it for orchestration decisions is redundant.
func determineModeFromServices() (string, error) {
	// Query knowledge service
	knowledgeResp, err := http.Get("http://knowledge:8084/state")
	if err != nil {
		log.Printf("Error querying knowledge service: %v\n", err)
		return "local", nil
	}
	defer knowledgeResp.Body.Close()

	var knowledgeState KnowledgeState
	if err := json.NewDecoder(knowledgeResp.Body).Decode(&knowledgeState); err != nil {
		log.Printf("Error parsing knowledge state: %v\n", err)
		return "local", nil
	}

	if knowledgeState.Fallback {
		return "local", nil
	}
	return "provider", nil
}

func orchestrate() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		mode, err := determineModeFromServices()
		if err != nil {
			log.Printf("Error determining mode: %v\n", err)
			continue
		}

		if err := setModeOnCore(mode); err != nil {
			log.Printf("Error setting mode on core: %v\n", err)
		}
	}
}

// healthHandler provides a health check endpoint
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"healthy": true}`))
}

func main() {
	go orchestrate()

	http.HandleFunc("/health", healthHandler)

	http.ListenAndServe(":8080", nil)
}
