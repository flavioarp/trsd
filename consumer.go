package main

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

var providerURL = "http://localhost:8081"
var fallbackMode = false

func callProvider() {

	resp, err := http.Get(providerURL + "/pdfa")

	if err != nil {
		fmt.Println("[CONSUMER] erro ao conectar no provider")
		enableFallback()
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Println("[CONSUMER] provider não suporta /pdfa")
		enableFallback()
		return
	}

	body, _ := io.ReadAll(resp.Body)

	fmt.Println("[CONSUMER] resposta do provider:", string(body))
}

func enableFallback() {

	if !fallbackMode {
		fmt.Println("################################")
		fmt.Println("[CONSUMER] ATIVANDO MODO FALLBACK")
		fmt.Println("################################")
	}

	fallbackMode = true
}

func localPDFA(w http.ResponseWriter, r *http.Request) {
	fmt.Println("[CONSUMER] executando geração LOCAL de PDF/A")
	w.Write([]byte("PDF/A gerado localmente (fallback)"))
}

func healthMonitor() {

	for {
		time.Sleep(5 * time.Second)

		if fallbackMode {
			fmt.Println("[CONSUMER] verificando se provider voltou...")

			resp, err := http.Get(providerURL + "/pdfa")

			if err == nil && resp.StatusCode == 200 {

				fmt.Println("********************************")
				fmt.Println("[CONSUMER] provider recuperado!")
				fmt.Println("********************************")

				fallbackMode = false
			}
		}
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	if fallbackMode {

		fmt.Println("[CONSUMER] usando fallback local")
		localPDFA(w, r)

		return
	}

	resp, err := http.Get(providerURL + "/pdfa")

	if err != nil || resp.StatusCode != 200 {

		enableFallback()
		localPDFA(w, r)

		return
	}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	fmt.Println("[CONSUMER] resposta do provider:", string(body))
	w.Write(body)
}

func main() {
	fmt.Println("================================")
	fmt.Println("[CONSUMER] iniciando serviço")
	fmt.Println("================================")

	http.HandleFunc("/pdfa", handler)

	go healthMonitor()

	http.ListenAndServe(":8080", nil)
}
