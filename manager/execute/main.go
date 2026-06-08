package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// Estrutura que representa a resposta do Knowledge Service
type KnowledgeState struct {
	Fallback        bool     `json:"fallback"`
	ActiveProviders []string `json:"active_providers"`
}

// Estrutura para enviar o comando de mudança de contexto para o Core
type ModeRequest struct {
	Mode      string   `json:"mode"`
	Providers []string `json:"providers"`
}

// Envia a decisão (Modo de Operação e Lista de Providers) para o sistema gerenciado (Core)
func setModeOnCore(state KnowledgeState) error {
	mode := "provider"
	if state.Fallback {
		mode = "local"
	}

	modeReq := ModeRequest{
		Mode:      mode,
		Providers: state.ActiveProviders,
	}
	modeBody, _ := json.Marshal(modeReq)

	req, err := http.NewRequest(http.MethodPost, "http://core:8082/mode", bytes.NewBuffer(modeBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Caller", "execute")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("core retornou status %d ao setar modo", resp.StatusCode)
	}

	log.Printf("modo do core atualizado para: %s | providers: %v\n", mode, state.ActiveProviders)
	return nil
}

// Consulta o serviço de conhecimento (Knowledge) para determinar as diretrizes
func determineStateFromKnowledge() (KnowledgeState, error) {
	var state KnowledgeState

	knowledgeResp, err := http.Get("http://knowledge:8084/state")
	if err != nil {
		log.Printf("Erro ao consultar serviço knowledge: %v\n", err)
		state.Fallback = true // Assume fallback por segurança se o knowledge cair
		return state, err
	}
	defer knowledgeResp.Body.Close()

	if err := json.NewDecoder(knowledgeResp.Body).Decode(&state); err != nil {
		log.Printf("Erro ao fazer parse do estado de knowledge: %v\n", err)
		state.Fallback = true
		return state, err
	}

	return state, nil
}

func execute() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		state, err := determineStateFromKnowledge()
		if err != nil {
			log.Printf("Aviso: Forçando fallback devido à falha de comunicação interna: %v\n", err)
		}

		if err := setModeOnCore(state); err != nil {
			log.Printf("Erro ao definir modo no core: %v\n", err)
		}
	}
}

// Endpoint de health check usado pelo core para validar se o manager está de pé
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"healthy": true}`))
}

func main() {
	go execute()

	http.HandleFunc("/health", healthHandler)

	log.Println("Starting Execute service no :8080...")
	http.ListenAndServe(":8080", nil)
}