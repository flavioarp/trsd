package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
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

const authorizedCaller = "execute"

const executeURL = "http://execute:8080/health"

var (
	currentMode = "provider"
	modeMutex   sync.RWMutex
)

func isExecuteHealthy() bool {
	client := &http.Client{Timeout: 500 * time.Millisecond}
	resp, err := client.Get(executeURL)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func getPDFAHandler(w http.ResponseWriter, r *http.Request) {
	if !isExecuteHealthy() {
		log.Println("Execute indisponível — forçando modo local")
		modeMutex.Lock()
		currentMode = "local"
		modeMutex.Unlock()
	}

	modeMutex.RLock()
	mode := currentMode
	modeMutex.RUnlock()

	if mode == "local" {
		w.Header().Set("Content-Type", "application/json")
		pdfResponse := PDFAResponse{
			Content: "PDF/A gerado localmente",
			Status:  "local",
		}
		json.NewEncoder(w).Encode(pdfResponse)
		return
	}

	// mode == "provider"
	client := &http.Client{
		Timeout: 700 * time.Millisecond,
	}

	resp, err := client.Get("http://provider:8081/pdfa")
	if err != nil || resp.StatusCode != http.StatusOK {
		// Se o provider falhar, retorne um PDF/A local, ainda que o modo seja "provider"
		w.Header().Set("Content-Type", "application/json")
		pdfResponse := PDFAResponse{
			Content: "PDF/A gerado localmente. O provider está indisponível ou sem suporte a PDF/A.",
			Status:  "fallback_error",
		}
		json.NewEncoder(w).Encode(pdfResponse)
		return
	}
	defer resp.Body.Close()

	// Passar o PDF/A do provider diretamente para o cliente, com base no status "provider"
	w.Header().Set("Content-Type", "application/octet-stream")
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "erro ao ler resposta do provider", http.StatusInternalServerError)
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

	// GET
	if r.Method == http.MethodGet {
		modeMutex.RLock()
		mode := currentMode
		modeMutex.RUnlock()

		json.NewEncoder(w).Encode(ModeResponse{Mode: mode})
		return
	}

	// POST
	if r.Method == http.MethodPost {
		if r.Header.Get("X-Caller") != authorizedCaller {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

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
	data, err := os.ReadFile("logo.txt")
	if err != nil {
		panic(err)
	}

	fmt.Println(string(data))

	http.HandleFunc("/pdfa", getPDFAHandler)
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/mode", modeHandler)

	fmt.Println("Starting PDF/A service...")

	err = http.ListenAndServe(":8082", nil)
	if err != nil {
		log.Fatal(err)
	}
}
