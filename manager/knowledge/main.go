package main

import (
	"encoding/json"
	"net/http"
	"sync"
)

type State struct {
	Fallback bool `json:"fallback"`
}

var (
	mu    sync.RWMutex
	state = State{Fallback: false}
)

func getState(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	defer mu.RUnlock()
	json.NewEncoder(w).Encode(state)
}

func setState(w http.ResponseWriter, r *http.Request) {
	var s State
	json.NewDecoder(r.Body).Decode(&s)

	mu.Lock()
	state = s
	mu.Unlock()

	w.WriteHeader(http.StatusOK)
}

func main() {
	http.HandleFunc("/state", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			getState(w, r)
		} else if r.Method == "POST" {
			setState(w, r)
		}
	})

	http.ListenAndServe(":8084", nil)
}
