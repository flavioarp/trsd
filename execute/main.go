package main

import (
	"encoding/json"
	"io"
	"net/http"
)

type State struct {
	Fallback bool `json:"fallback"`
}

func handler(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get("http://knowledge:8084/state")
	if err != nil {
		http.Error(w, "knowledge down", 500)
		return
	}
	defer resp.Body.Close()

	var s State
	json.NewDecoder(resp.Body).Decode(&s)

	if s.Fallback {
		w.Write([]byte("PDF/A gerado localmente (fallback)"))
		return
	}

	resp2, err := http.Get("http://provider:8081/pdfa")
	if err != nil || resp2.StatusCode != 200 {
		w.Write([]byte("fallback após erro"))
		return
	}
	defer resp2.Body.Close()

	body, _ := io.ReadAll(resp2.Body)
	w.Write(body)
}

func main() {
	http.HandleFunc("/pdfa", handler)
	http.ListenAndServe(":8080", nil)
}
