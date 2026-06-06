package main

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type Input struct {
	Fallback        bool     `json:"fallback"`
	ActiveProviders []string `json:"active_providers"`
}

func plan(w http.ResponseWriter, r *http.Request) {
	var in Input
	json.NewDecoder(r.Body).Decode(&in)

	body, _ := json.Marshal(in)

	http.Post("http://knowledge:8084/state", "application/json", bytes.NewBuffer(body))
}

func main() {
	http.HandleFunc("/plan", plan)
	http.ListenAndServe(":8083", nil)
}