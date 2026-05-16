package main

import (
	"fmt"
	"net/http"
	"os"
)

var service = "[PROVIDER]"

var enablePDFA = false

func health(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("O provider está saudável."))
}

func pdfa(w http.ResponseWriter, r *http.Request) {
	if !enablePDFA {
		res := "Suporte a PDF/A está desativado."
		http.Error(w, res, http.StatusServiceUnavailable)
		fmt.Println(service, res)

		return
	}

	w.Write([]byte("PDF/A gerado pelo provider."))
}

func main() {
	if os.Getenv("ENABLE_PDFA") == "1" {
		enablePDFA = true
	} else if len(os.Args) > 1 && os.Args[1] == "pdfa" {
		enablePDFA = true
	}

	var port = "8081"

	fmt.Println(service, "Iniciado em", port, "/tcp.", " PDF/A:", enablePDFA)

	http.HandleFunc("/health", health)
	http.HandleFunc("/pdfa", pdfa)

	http.ListenAndServe(":"+port, nil)
}
