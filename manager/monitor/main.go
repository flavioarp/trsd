// trsd-main/manager/monitor/main.go
package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

//Lista de Providers
var knownProviders = []string{
	"http://provider1:8081",
	"http://provider2:8081",
}

func checkProvider(url string) bool {
	client := http.Client{Timeout: 1 * time.Second}
	resp, err := client.Get(url + "/health")
	if err != nil || resp.StatusCode != 200 {
		return false
	}
	defer resp.Body.Close()

	var status map[string]bool
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return false
	}
	// Só retorna true se estiver saudável e com PDF/A ativo
	return status["healthy"] && status["pdfa_enabled"]
}

func main() {
	for {
		var activeProviders []string

		for _, p := range knownProviders {
			if checkProvider(p) {
				activeProviders = append(activeProviders, p)
			}
		}

		body, _ := json.Marshal(map[string]interface{}{
			"active_providers": activeProviders,
		})

		http.Post("http://analyze:8082/analyze", "application/json", bytes.NewBuffer(body))
		time.Sleep(1 * time.Second)
	}
}