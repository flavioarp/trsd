// trsd-main/manager/analyze/main.go
package main

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type Input struct {
	ActiveProviders []string `json:"active_providers"`
}

func analyze(w http.ResponseWriter, r *http.Request) {
	var in Input
	json.NewDecoder(r.Body).Decode(&in)

	// Fallback ativado apenas se nenhum provider estiver ativo e habilitado
	fallback := len(in.ActiveProviders) == 0

	body, _ := json.Marshal(map[string]interface{}{
		"fallback":         fallback,
		"active_providers": in.ActiveProviders,
	})

	http.Post("http://plan:8083/plan", "application/json", bytes.NewBuffer(body))
}

func main() {
	http.HandleFunc("/analyze", analyze)
	http.ListenAndServe(":8082", nil)
}