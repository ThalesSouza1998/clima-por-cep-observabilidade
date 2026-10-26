package servicob

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/ThalesSouza1998/clima-por-cep-observabilidade/internal/clima"
	"github.com/ThalesSouza1998/clima-por-cep-observabilidade/internal/viacep"
)

type buscaFalsa struct {
	cidade string
	err    error
}

func (b buscaFalsa) Cidade(context.Context, string) (string, error) {
	return b.cidade, b.err
}

type termometroFalso struct {
	celsius float64
	err     error
}

func (t termometroFalso) TemperaturaAtual(context.Context, string) (float64, error) {
	return t.celsius, t.err
}

func TestHandlerClima(t *testing.T) {
	casos := []struct {
		nome       string
		corpo      string
		busca      buscaFalsa
		termometro termometroFalso
		querStatus int
		querCorpo  map[string]any
	}{
		{
			nome:       "sucesso",
			corpo:      `{"cep":"29902555"}`,
			busca:      buscaFalsa{cidade: "Linhares"},
			termometro: termometroFalso{celsius: 23},
			querStatus: http.StatusOK,
			querCorpo: map[string]any{
				"city":   "Linhares",
				"temp_C": 23.0,
				"temp_F": 73.4,
				"temp_K": 296.15,
			},
		},
		{
			nome:       "cep inválido",
			corpo:      `{"cep":"123"}`,
			querStatus: http.StatusUnprocessableEntity,
			querCorpo:  map[string]any{"error": "invalid zipcode"},
		},
		{
			nome:       "cep não é string",
			corpo:      `{"cep":29902555}`,
			querStatus: http.StatusUnprocessableEntity,
			querCorpo:  map[string]any{"error": "invalid zipcode"},
		},
		{
			nome:       "cep não encontrado",
			corpo:      `{"cep":"99999999"}`,
			busca:      buscaFalsa{err: viacep.ErrCEPDesconhecido},
			querStatus: http.StatusNotFound,
			querCorpo:  map[string]any{"error": "can not find zipcode"},
		},
		{
			nome:       "falha na weatherapi",
			corpo:      `{"cep":"29902555"}`,
			busca:      buscaFalsa{cidade: "Linhares"},
			termometro: termometroFalso{err: errors.New("falha")},
			querStatus: http.StatusInternalServerError,
			querCorpo:  map[string]any{"error": "internal error"},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			handler := Novo(clima.NovoServico(caso.busca, caso.termometro))
			requisicao := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(caso.corpo))
			gravador := httptest.NewRecorder()

			handler.Clima(gravador, requisicao)

			if gravador.Code != caso.querStatus {
				t.Fatalf("status = %d; queria %d", gravador.Code, caso.querStatus)
			}

			if tipo := gravador.Header().Get("Content-Type"); tipo != "application/json" {
				t.Errorf("Content-Type = %q; queria application/json", tipo)
			}

			var corpo map[string]any
			if err := json.Unmarshal(gravador.Body.Bytes(), &corpo); err != nil {
				t.Fatalf("resposta não é um JSON válido: %v", err)
			}

			if !reflect.DeepEqual(corpo, caso.querCorpo) {
				t.Errorf("corpo = %v; queria %v", corpo, caso.querCorpo)
			}
		})
	}
}
