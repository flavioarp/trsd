package main

import (
	"fmt"
	"net/http"
	"os"
)

var enablePDFA bool

func health(w http.ResponseWriter, r *http.Request) {
	fmt.Println("[PROVIDER] healthcheck recebido")
	w.Write([]byte("ok"))
}

func pdfa(w http.ResponseWriter, r *http.Request) {
	if !enablePDFA {
		fmt.Println("[PROVIDER] /pdfa solicitado mas não suportado")
		http.NotFound(w, r)
		return
	}

	fmt.Println("[PROVIDER] gerando PDF/A")
	w.Write([]byte("PDF/A gerado pelo provider"))
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "pdfa" {
		enablePDFA = true
	}

	fmt.Println("================================")
	fmt.Println("[PROVIDER] iniciando serviço")
	fmt.Println("[PROVIDER] PDF/A habilitado:", enablePDFA)
	fmt.Println("================================")

	http.HandleFunc("/health", health)

	if enablePDFA {
		http.HandleFunc("/pdfa", pdfa)
	}

	http.ListenAndServe(":8081", nil)
}
