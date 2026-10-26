// Package weatherapi consulta a temperatura atual na WeatherAPI.
package weatherapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"github.com/ThalesSouza1998/clima-por-cep-observabilidade/internal/telemetria"
)

type Cliente struct {
	cliente      *http.Client
	chave        string
	enderecoBase string
}

func Novo(chave string) *Cliente {
	return &Cliente{
		cliente:      &http.Client{Timeout: 10 * time.Second},
		chave:        chave,
		enderecoBase: "https://api.weatherapi.com",
	}
}

type resposta struct {
	Current struct {
		Celsius float64 `json:"temp_c"`
	} `json:"current"`
}

func (c *Cliente) TemperaturaAtual(ctx context.Context, cidade string) (float64, error) {
	ctx, span := otel.Tracer("weatherapi").Start(ctx, "busca-temperatura")
	defer span.End()
	span.SetAttributes(attribute.String("cidade", cidade))

	parametros := url.Values{}
	parametros.Set("key", c.chave)
	parametros.Set("q", cidade)
	parametros.Set("aqi", "no")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.enderecoBase+"/v1/current.json?"+parametros.Encode(), nil)
	if err != nil {
		telemetria.RegistrarErro(span, err)
		return 0, fmt.Errorf("montando a requisição da WeatherAPI: %w", err)
	}

	resp, err := c.cliente.Do(req)
	if err != nil {
		err = fmt.Errorf("consultando a WeatherAPI: %w", semURL(err))
		telemetria.RegistrarErro(span, err)
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("WeatherAPI respondeu %s", resp.Status)
		telemetria.RegistrarErro(span, err)
		return 0, err
	}

	var dados resposta
	if err := json.NewDecoder(resp.Body).Decode(&dados); err != nil {
		telemetria.RegistrarErro(span, err)
		return 0, fmt.Errorf("lendo a resposta da WeatherAPI: %w", err)
	}

	span.SetAttributes(attribute.Float64("temperatura_celsius", dados.Current.Celsius))
	return dados.Current.Celsius, nil
}

// semURL remove a URL do erro para não levar a chave da API para logs e traces.
func semURL(err error) error {
	for {
		var erroDeURL *url.Error
		if !errors.As(err, &erroDeURL) {
			return err
		}
		if erroDeURL.Err == nil {
			return errors.New("falha de rede")
		}
		err = erroDeURL.Err
	}
}
