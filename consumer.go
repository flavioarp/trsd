package main

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

const providerURL = "http://localhost:8081"

type Knowledge struct {
	mu           sync.RWMutex
	fallbackMode bool
}

var k = &Knowledge{}

func monitor() bool {
	resp, err := http.Get(providerURL + "/pdfa")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}

func analyze(ok bool) {
	k.mu.Lock()
	defer k.mu.Unlock()

	if !ok && !k.fallbackMode {
		fmt.Println("(A) Provider indisponível -> ativando fallback")
		k.fallbackMode = true
	}

	if ok && k.fallbackMode {
		fmt.Println("(A) Provider recuperado -> desativando fallback")
		k.fallbackMode = false
	}
}

func plan() bool {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.fallbackMode
}

func execute(w http.ResponseWriter, fallback bool) {
	if fallback {
		localPDFA(w)
		return
	}

	resp, err := http.Get(providerURL + "/pdfa")
	if err != nil || resp.StatusCode == 503 {
		fmt.Println("(E) falha -> fallback")
		localPDFA(w)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	w.Write(body)
}

// Knowledge action
func localPDFA(w http.ResponseWriter) {
	fmt.Println("(K) Gerando PDF/A local")
	w.Write([]byte("PDF/A gerado localmente"))
}

func mape() {
	for {
		time.Sleep(5 * time.Second)
		ok := monitor()
		analyze(ok)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	fallback := plan()
	execute(w, fallback)
}

func main() {
	fmt.Println("Consumer iniciado em 8080/tcp")

	http.HandleFunc("/pdfa", handler)

	go mape()

	http.ListenAndServe(":8080", nil)
}
