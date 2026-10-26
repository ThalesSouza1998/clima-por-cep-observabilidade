// Package viacep consulta o ViaCEP para descobrir a cidade de um CEP.
package viacep

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"github.com/ThalesSouza1998/clima-por-cep-observabilidade/internal/telemetria"
)

var ErrCEPDesconhecido = errors.New("cep não existe na base do ViaCEP")

type Cliente struct {
	cliente      *http.Client
	enderecoBase string
}

func Novo() *Cliente {
	return &Cliente{
		cliente:      &http.Client{Timeout: 10 * time.Second},
		enderecoBase: "https://viacep.com.br",
	}
}

// Quando o CEP não existe o ViaCEP devolve {"erro": true}, sem localidade.
type endereco struct {
	Localidade string `json:"localidade"`
}

func (c *Cliente) Cidade(ctx context.Context, valor string) (string, error) {
	ctx, span := otel.Tracer("viacep").Start(ctx, "busca-cep")
	defer span.End()
	span.SetAttributes(attribute.String("cep", valor))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.enderecoBase+"/ws/"+valor+"/json/", nil)
	if err != nil {
		telemetria.RegistrarErro(span, err)
		return "", fmt.Errorf("montando a requisição do ViaCEP: %w", err)
	}

	resp, err := c.cliente.Do(req)
	if err != nil {
		telemetria.RegistrarErro(span, err)
		return "", fmt.Errorf("consultando o ViaCEP: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("ViaCEP respondeu %s", resp.Status)
		telemetria.RegistrarErro(span, err)
		return "", err
	}

	var dados endereco
	if err := json.NewDecoder(resp.Body).Decode(&dados); err != nil {
		telemetria.RegistrarErro(span, err)
		return "", fmt.Errorf("lendo a resposta do ViaCEP: %w", err)
	}

	if dados.Localidade == "" {
		telemetria.RegistrarErro(span, ErrCEPDesconhecido)
		return "", ErrCEPDesconhecido
	}

	span.SetAttributes(attribute.String("cidade", dados.Localidade))
	return dados.Localidade, nil
}
