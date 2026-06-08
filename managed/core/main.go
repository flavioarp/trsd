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

type ModeRequest struct {
	Mode      string   `json:"mode"` // "local" or "provider"
	Providers []string `json:"providers"`
}

type ModeResponse struct {
	Mode      string   `json:"mode"`
	Providers []string `json:"providers"`
}

type PDFAResponse struct {
	Content string `json:"content"`
	Status  string `json:"status"`
}

const authorizedCaller = "execute"
const executeURL = "http://execute:8080/health"

var (
	currentMode     = "provider"
	activeProviders = []string{}
	providerIndex   = 0 // Usado para rotacionar no Load Balancer (Round-Robin)
	modeMutex       sync.RWMutex
)

func isExecuteHealthy() bool {
	client := &http.Client{Timeout: 500 * time.Millisecond}
	resp, err := client.Get(executeURL)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func getPDFAHandler(w http.ResponseWriter, r *http.Request) {
	modeMutex.Lock()
	// Verificação de segurança: Se o próprio gerente (MAPE) tiver caído, assume independência
	if !isExecuteHealthy() {
		log.Println("Execute indisponível — forçando modo local")
		currentMode = "local"
	}

	mode := currentMode
	providers := activeProviders

	// Lógica de Load Balancing (Round-Robin)
	var selectedProvider string
	if len(providers) > 0 {
		selectedProvider = providers[providerIndex%len(providers)]
		providerIndex++
	}
	modeMutex.Unlock()

	// Entra em contingência direta se o modo for local ou se a lista de providers for enviada vazia
	if mode == "local" || len(providers) == 0 {
		w.Header().Set("Content-Type", "application/json")
		pdfResponse := PDFAResponse{
			Content: "PDF/A gerado localmente (Fallback ou Nenhum Provider Ativo)",
			Status:  "local",
		}
		json.NewEncoder(w).Encode(pdfResponse)
		return
	}

	// Tenta acessar o provider que foi selecionado pelo Round-Robin
	client := &http.Client{
		Timeout: 700 * time.Millisecond,
	}

	resp, err := client.Get(selectedProvider + "/pdfa")
	if err != nil || resp.StatusCode != http.StatusOK {
		// Tolerância a falhas na execução: Mesmo que o provider fosse válido 1 segundo atrás, 
		// caso ele caia nesse exato momento, aciona contingência na mesma requisição.
		w.Header().Set("Content-Type", "application/json")
		pdfResponse := PDFAResponse{
			Content: fmt.Sprintf("PDF/A gerado localmente. O provider %s está indisponível neste momento.", selectedProvider),
			Status:  "fallback_error",
		}
		json.NewEncoder(w).Encode(pdfResponse)
		return
	}
	defer resp.Body.Close()

	// Passar o PDF/A com sucesso, direto para o cliente do Core
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

	// GET: Retorna o status atual
	if r.Method == http.MethodGet {
		modeMutex.RLock()
		mode := currentMode
		providers := activeProviders
		modeMutex.RUnlock()

		json.NewEncoder(w).Encode(ModeResponse{Mode: mode, Providers: providers})
		return
	}

	// POST: Usado pelo serviço Manager (execute) para alterar as políticas dinâmicas
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
		activeProviders = req.Providers
		modeMutex.Unlock()

		json.NewEncoder(w).Encode(ModeResponse{Mode: req.Mode, Providers: req.Providers})
		return
	}

	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func main() {
	// Logo é opcional. Previne panic caso rodando de outro diretório ou se faltando o txt
	data, err := os.ReadFile("logo.txt")
	if err == nil {
		fmt.Println(string(data))
	}

	http.HandleFunc("/pdfa", getPDFAHandler)
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/mode", modeHandler)

	fmt.Println("Starting PDF/A Core service em :8082...")

	err = http.ListenAndServe(":8082", nil)
	if err != nil {
		log.Fatal(err)
	}
}