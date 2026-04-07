package main

import (
	"fmt"
	"net/http"
	"os"
)

var enablePDFA = false

func health(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}

func pdfa(w http.ResponseWriter, r *http.Request) {
	if !enablePDFA {
		res := "Consumer solicitou o recurso /pdfa, porém está desativado."
		http.Error(w, res, http.StatusServiceUnavailable)
		fmt.Println(res)
		return
	}
	w.Write([]byte("Consumer solicitou o recurso /pdfa... PDF/A gerado pelo provider!"))
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "pdfa" {
		enablePDFA = true
	}

	fmt.Println("[PROVIDER] iniciado em 8081/tcp | PDF/A:", enablePDFA)

	http.HandleFunc("/health", health)
	http.HandleFunc("/pdfa", pdfa)

	http.ListenAndServe(":8081", nil)
}
