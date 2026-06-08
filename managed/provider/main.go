package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
)

var (
	service    = "[PROVIDER]"
	enablePDFA = false
	mu         sync.RWMutex // Protege a variável contra leituras/escritas simultâneas
)

func health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	mu.RLock()
	status := enablePDFA
	mu.RUnlock()

	json.NewEncoder(w).Encode(map[string]bool{
		"healthy":      true,
		"pdfa_enabled": status,
	})
}

func pdfa(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	status := enablePDFA
	mu.RUnlock()

	if !status {
		http.Error(w, "Suporte a PDF/A está desativado.", http.StatusServiceUnavailable)
		return
	}
	w.Write([]byte(fmt.Sprintf("PDF/A gerado pelo %s.", os.Getenv("HOSTNAME"))))
}

// NOVA ROTA: Permite ativar/desativar em tempo real
func toggleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	var config struct {
		EnablePDFA bool `json:"enable_pdfa"`
	}

	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	// Altera a variável global com segurança
	mu.Lock()
	enablePDFA = config.EnablePDFA
	mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":      "Configuração atualizada com sucesso",
		"pdfa_enabled": config.EnablePDFA,
	})
}

func main() {
	if os.Getenv("ENABLE_PDFA") == "1" || os.Getenv("ENABLE_PDFA") == "true" {
		enablePDFA = true
	}

	port := "8081"
	fmt.Println(service, "Iniciado em", port, "/tcp.", " PDF/A:", enablePDFA)

	http.HandleFunc("/health", health)
	http.HandleFunc("/pdfa", pdfa)
	http.HandleFunc("/toggle", toggleConfig) // Registrando a nova rota

	http.ListenAndServe(":"+port, nil)
}