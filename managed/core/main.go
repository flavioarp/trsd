package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

type State struct {
	Fallback bool `json:"fallback"`
}

type ModeRequest struct {
	Mode string `json:"mode"` // "local" or "provider"
}

type ModeResponse struct {
	Mode string `json:"mode"`
}

type PDFAResponse struct {
	Content string `json:"content"`
	Status  string `json:"status"`
}

var (
	currentMode = "provider"
	modeMutex   sync.RWMutex
)

// getPDFAHandler handles the PDF/A generation with fallback mechanism
// It respects the mode set by the execute service and decides whether to use
// local generation or delegate to the provider service
func getPDFAHandler(w http.ResponseWriter, r *http.Request) {
	modeMutex.RLock()
	mode := currentMode
	modeMutex.RUnlock()

	// If mode is "local", always use local generation
	if mode == "local" {
		w.Header().Set("Content-Type", "application/json")
		pdfResponse := PDFAResponse{
			Content: "PDF/A gerado localmente",
			Status:  "local",
		}
		json.NewEncoder(w).Encode(pdfResponse)
		return
	}

	// If mode is "provider", try to get PDF/A from provider service
	client := &http.Client{
		Timeout: 700 * time.Millisecond,
	}

	resp, err := client.Get("http://provider:8081/pdfa")
	if err != nil || resp.StatusCode != http.StatusOK {
		// If provider call fails or times out, fall back to local generation
		w.Header().Set("Content-Type", "application/json")
		pdfResponse := PDFAResponse{
			Content: "PDF/A gerado localmente. O provider está indisponível ou sem suporte a PDF/A.",
			Status:  "fallback_error",
		}
		json.NewEncoder(w).Encode(pdfResponse)
		return
	}
	defer resp.Body.Close()

	// Provider service is available and healthy, return its response
	w.Header().Set("Content-Type", "application/octet-stream")
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "failed to read provider response", http.StatusInternalServerError)
		return
	}

	w.Write(body)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"healthy": true})
}

func modeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodGet {
		// GET: return current mode
		modeMutex.RLock()
		mode := currentMode
		modeMutex.RUnlock()

		json.NewEncoder(w).Encode(ModeResponse{Mode: mode})
		return
	}

	if r.Method == http.MethodPost {
		// POST: set new mode
		var req ModeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if req.Mode != "local" && req.Mode != "provider" {
			http.Error(w, "invalid mode: must be 'local' or 'provider'", http.StatusBadRequest)
			return
		}

		modeMutex.Lock()
		currentMode = req.Mode
		modeMutex.Unlock()

		json.NewEncoder(w).Encode(ModeResponse{Mode: req.Mode})
		return
	}

	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func main() {
	data, err := os.ReadFile("pdf-a-daptive.txt")
	if err != nil {
		panic(err)
	}

	content := string(data)

	fmt.Println(content)

	fmt.Println("Starting PDF/A service...")

	http.HandleFunc("/pdfa", getPDFAHandler)

	http.HandleFunc("/health", healthHandler)

	// Mode control endpoint (GET to query, POST to set)
	http.HandleFunc("/mode", modeHandler)

	http.ListenAndServe(":8082", nil)
}
