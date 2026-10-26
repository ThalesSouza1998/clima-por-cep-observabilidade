// Package servicoa implementa a porta de entrada do sistema.
package servicoa

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/ThalesSouza1998/clima-por-cep-observabilidade/internal/cep"
	"github.com/ThalesSouza1998/clima-por-cep-observabilidade/internal/httpjson"
)

const tamanhoMaximoCorpo = 1 << 20

type Handler struct {
	cliente *http.Client
	destino string
}

func Novo(destino string) *Handler {
	return &Handler{
		cliente: &http.Client{
			Timeout:   15 * time.Second,
			Transport: otelhttp.NewTransport(http.DefaultTransport),
		},
		destino: destino,
	}
}

type requisicao struct {
	CEP any `json:"cep"`
}

type encaminhamento struct {
	CEP string `json:"cep"`
}

func (h *Handler) Clima(w http.ResponseWriter, r *http.Request) {
	var pedido requisicao
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, tamanhoMaximoCorpo)).Decode(&pedido); err != nil {
		httpjson.Erro(w, http.StatusUnprocessableEntity, "invalid zipcode")
		return
	}

	valor, ok := pedido.CEP.(string)
	if !ok || !cep.Valido(valor) {
		httpjson.Erro(w, http.StatusUnprocessableEntity, "invalid zipcode")
		return
	}

	corpo, err := json.Marshal(encaminhamento{CEP: valor})
	if err != nil {
		log.Printf("montando o corpo para o serviço B: %v", err)
		httpjson.Erro(w, http.StatusInternalServerError, "internal error")
		return
	}

	requisicaoB, err := http.NewRequestWithContext(r.Context(), http.MethodPost, h.destino, bytes.NewReader(corpo))
	if err != nil {
		log.Printf("montando a requisição para o serviço B: %v", err)
		httpjson.Erro(w, http.StatusInternalServerError, "internal error")
		return
	}
	requisicaoB.Header.Set("Content-Type", "application/json")

	respostaB, err := h.cliente.Do(requisicaoB)
	if err != nil {
		log.Printf("serviço B indisponível: %v", err)
		httpjson.Erro(w, http.StatusBadGateway, "service unavailable")
		return
	}
	defer respostaB.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(respostaB.StatusCode)

	if _, err := io.Copy(w, respostaB.Body); err != nil {
		log.Printf("copiando a resposta do serviço B: %v", err)
	}
}
