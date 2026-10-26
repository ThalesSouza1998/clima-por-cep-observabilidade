package viacep

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestCidadeCriaSpanDaBusca(t *testing.T) {
	servidor := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"localidade":"Linhares"}`))
	}))
	defer servidor.Close()

	exportador := tracetest.NewInMemoryExporter()
	provedor := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exportador))
	anterior := otel.GetTracerProvider()
	otel.SetTracerProvider(provedor)
	defer otel.SetTracerProvider(anterior)

	cliente := &Cliente{cliente: servidor.Client(), enderecoBase: servidor.URL}

	cidade, err := cliente.Cidade(context.Background(), "29902555")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if cidade != "Linhares" {
		t.Fatalf("cidade = %q; queria Linhares", cidade)
	}

	spans := exportador.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("spans = %d; queria 1", len(spans))
	}
	if spans[0].Name != "busca-cep" {
		t.Errorf("nome do span = %q; queria busca-cep", spans[0].Name)
	}
	if valor := valorDoAtributo(spans[0].Attributes, "cep"); valor != "29902555" {
		t.Errorf("atributo cep = %q", valor)
	}
	if valor := valorDoAtributo(spans[0].Attributes, "cidade"); valor != "Linhares" {
		t.Errorf("atributo cidade = %q", valor)
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
