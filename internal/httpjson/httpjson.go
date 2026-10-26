// Package httpjson padroniza as respostas JSON dos dois serviços.
package httpjson

import (
	"encoding/json"
	"log"
	"net/http"
)

func Escrever(w http.ResponseWriter, status int, corpo any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(corpo); err != nil {
		log.Printf("erro ao escrever a resposta: %v", err)
	}
}

func Erro(w http.ResponseWriter, status int, mensagem string) {
	Escrever(w, status, map[string]string{"error": mensagem})
}
