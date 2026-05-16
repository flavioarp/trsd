package main

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type Input struct {
	ProviderOK bool `json:"provider_ok"`
}

func analyze(w http.ResponseWriter, r *http.Request) {
	var in Input
	json.NewDecoder(r.Body).Decode(&in)

	fallback := !in.ProviderOK

	body, _ := json.Marshal(map[string]bool{
		"fallback": fallback,
	})

	http.Post("http://plan:8083/plan", "application/json", bytes.NewBuffer(body))
}

func main() {
	http.HandleFunc("/analyze", analyze)
	http.ListenAndServe(":8082", nil)
}
