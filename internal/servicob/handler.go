// Package servicob implementa a orquestração entre ViaCEP e WeatherAPI.
package servicob

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/ThalesSouza1998/clima-por-cep-observabilidade/internal/cep"
	"github.com/ThalesSouza1998/clima-por-cep-observabilidade/internal/clima"
	"github.com/ThalesSouza1998/clima-por-cep-observabilidade/internal/httpjson"
)

const tamanhoMaximoCorpo = 1 << 20

type Handler struct {
	servico *clima.Servico
}

func Novo(servico *clima.Servico) *Handler {
	return &Handler{servico: servico}
}

type requisicao struct {
	CEP any `json:"cep"`
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

	previsao, err := h.servico.PorCEP(r.Context(), valor)
	if err != nil {
		switch {
		case errors.Is(err, clima.ErrCEPinvalido):
			httpjson.Erro(w, http.StatusUnprocessableEntity, "invalid zipcode")
		case errors.Is(err, clima.ErrCEPNaoEncontrado):
			httpjson.Erro(w, http.StatusNotFound, "can not find zipcode")
		default:
			log.Printf("erro ao consultar o clima do CEP %s: %v", valor, err)
			httpjson.Erro(w, http.StatusInternalServerError, "internal error")
		}
		return
	}

	httpjson.Escrever(w, http.StatusOK, previsao)
}
