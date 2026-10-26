package servicoa

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestHandlerEncaminhaParaServicoB(t *testing.T) {
	recebidos := make(chan string, 1)

	servidorB := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		corpo, _ := io.ReadAll(r.Body)
		recebidos <- string(corpo)

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"city":"Linhares","temp_C":23,"temp_F":73.4,"temp_K":296.15}`))
	}))
	defer servidorB.Close()

	requisicao := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"cep":"29902555"}`))
	gravador := httptest.NewRecorder()

	Novo(servidorB.URL).Clima(gravador, requisicao)

	if gravador.Code != http.StatusOK {
		t.Fatalf("status = %d; queria %d", gravador.Code, http.StatusOK)
	}

	select {
	case corpo := <-recebidos:
		if corpo != `{"cep":"29902555"}` {
			t.Errorf("corpo enviado ao serviço B = %s", corpo)
		}
	default:
		t.Fatal("o serviço B não recebeu a requisição")
	}

	var resposta map[string]any
	if err := json.Unmarshal(gravador.Body.Bytes(), &resposta); err != nil {
		t.Fatalf("resposta não é um JSON válido: %v", err)
	}
	if resposta["city"] != "Linhares" {
		t.Errorf("city = %v; queria Linhares", resposta["city"])
	}
}

func TestHandlerRejeitaCEPinvalido(t *testing.T) {
	var chamadas atomic.Int32

	servidorB := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		chamadas.Add(1)
	}))
	defer servidorB.Close()

	handler := Novo(servidorB.URL)

	casos := []struct {
		nome  string
		corpo string
	}{
		{nome: "cep curto", corpo: `{"cep":"123"}`},
		{nome: "cep com letras", corpo: `{"cep":"0100100a"}`},
		{nome: "cep numérico", corpo: `{"cep":29902555}`},
		{nome: "sem cep", corpo: `{}`},
		{nome: "json inválido", corpo: `{`},
		{nome: "corpo vazio", corpo: ``},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			requisicao := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(caso.corpo))
			gravador := httptest.NewRecorder()

			handler.Clima(gravador, requisicao)

			if gravador.Code != http.StatusUnprocessableEntity {
				t.Fatalf("status = %d; queria %d", gravador.Code, http.StatusUnprocessableEntity)
			}

			var resposta map[string]string
			if err := json.Unmarshal(gravador.Body.Bytes(), &resposta); err != nil {
				t.Fatalf("resposta não é um JSON válido: %v", err)
			}
			if resposta["error"] != "invalid zipcode" {
				t.Errorf("mensagem = %q; queria invalid zipcode", resposta["error"])
			}
		})
	}

	if chamadas.Load() != 0 {
		t.Errorf("o serviço B foi chamado %d vez(es) para CEPs inválidos", chamadas.Load())
	}
}

func TestHandlerServicoBIndisponivel(t *testing.T) {
	requisicao := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"cep":"29902555"}`))
	gravador := httptest.NewRecorder()

	Novo("http://127.0.0.1:1").Clima(gravador, requisicao)

	if gravador.Code != http.StatusBadGateway {
		t.Fatalf("status = %d; queria %d", gravador.Code, http.StatusBadGateway)
	}
}

func TestHandlerPropagaTraceContext(t *testing.T) {
	exportador := tracetest.NewInMemoryExporter()
	provedor := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exportador))
	provedorAnterior := otel.GetTracerProvider()
	otel.SetTracerProvider(provedor)
	defer otel.SetTracerProvider(provedorAnterior)

	propagadorAnterior := otel.GetTextMapPropagator()
	otel.SetTextMapPropagator(propagation.TraceContext{})
	defer otel.SetTextMapPropagator(propagadorAnterior)

	cabecalhos := make(chan string, 1)
	servidorB := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cabecalhos <- r.Header.Get("traceparent")
		w.WriteHeader(http.StatusOK)
	}))
	defer servidorB.Close()

	requisicao := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"cep":"29902555"}`))
	gravador := httptest.NewRecorder()

	Novo(servidorB.URL).Clima(gravador, requisicao)

	select {
	case traceparent := <-cabecalhos:
		if traceparent == "" {
			t.Fatal("traceparent não foi enviado ao serviço B")
		}
	case <-time.After(time.Second):
		t.Fatal("o serviço B não recebeu a requisição")
	}
}
