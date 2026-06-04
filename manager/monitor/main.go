package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

func checkProvider() bool {
	resp, err := http.Get("http://provider:8081/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}

func main() {
	for {
		ok := checkProvider()

		body, _ := json.Marshal(map[string]bool{
			"provider_ok": ok,
		})

		http.Post("http://analyze:8082/analyze", "application/json", bytes.NewBuffer(body))

		time.Sleep(1 * time.Second)
	}
}
