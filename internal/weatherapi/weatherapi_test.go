package weatherapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestTemperaturaAtualCriaSpanDaBusca(t *testing.T) {
	servidor := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"current":{"temp_c":23}}`))
	}))
	defer servidor.Close()

	exportador := tracetest.NewInMemoryExporter()
	provedor := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exportador))
	anterior := otel.GetTracerProvider()
	otel.SetTracerProvider(provedor)
	defer otel.SetTracerProvider(anterior)

	cliente := Novo("chave-de-teste")
	cliente.cliente = servidor.Client()
	cliente.enderecoBase = servidor.URL

	temperatura, err := cliente.TemperaturaAtual(context.Background(), "Linhares")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if temperatura != 23 {
		t.Fatalf("temperatura = %v; queria 23", temperatura)
	}

	spans := exportador.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("spans = %d; queria 1", len(spans))
	}
	if spans[0].Name != "busca-temperatura" {
		t.Errorf("nome do span = %q; queria busca-temperatura", spans[0].Name)
	}
	if valor := valorDoAtributo(spans[0].Attributes, "cidade"); valor != "Linhares" {
		t.Errorf("atributo cidade = %q", valor)
	}
	if valor := valorDoAtributoNumerico(spans[0].Attributes, "temperatura_celsius"); valor != 23 {
		t.Errorf("atributo temperatura_celsius = %v", valor)
	}
}

func valorDoAtributo(atributos []attribute.KeyValue, chave string) string {
	for _, atributo := range atributos {
		if string(atributo.Key) == chave {
			return atributo.Value.AsString()
		}
	}
	return ""
}

func valorDoAtributoNumerico(atributos []attribute.KeyValue, chave string) float64 {
	for _, atributo := range atributos {
		if string(atributo.Key) == chave {
			return atributo.Value.AsFloat64()
		}
	}
	return 0
}

type transporteComErro struct{}

func (transporteComErro) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, &url.Error{
		Op:  "Get",
		URL: "https://api.weatherapi.com/v1/current.json?key=segredo",
		Err: errors.New("falha de rede"),
	}
}

func TestErroDeRedeNaoVazaAChave(t *testing.T) {
	cliente := Novo("segredo")
	cliente.cliente = &http.Client{Transport: transporteComErro{}}

	_, err := cliente.TemperaturaAtual(context.Background(), "Linhares")
	if err == nil {
		t.Fatal("esperava um erro")
	}
	if strings.Contains(err.Error(), "segredo") {
		t.Errorf("o erro vazou a chave da API: %v", err)
	}
}
