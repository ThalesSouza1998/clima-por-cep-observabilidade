// Package telemetria configura o OpenTelemetry dos dois serviços.
package telemetria

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// Iniciar configura o TracerProvider e o exportador OTLP. A função devolvida
// deve ser chamada no encerramento do processo para não perder spans.
func Iniciar(ctx context.Context, nomeServico string) (func(context.Context) error, error) {
	endereco := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endereco == "" {
		endereco = "http://localhost:4318"
	}
	if !strings.HasPrefix(endereco, "http://") && !strings.HasPrefix(endereco, "https://") {
		endereco = "http://" + endereco
	}

	destino, err := url.Parse(endereco)
	if err != nil {
		return nil, fmt.Errorf("interpretando OTEL_EXPORTER_OTLP_ENDPOINT: %w", err)
	}
	if destino.Path == "" || destino.Path == "/" {
		destino.Path = "/v1/traces"
	}

	opcoes := []otlptracehttp.Option{otlptracehttp.WithEndpointURL(destino.String())}
	if destino.Scheme == "http" {
		opcoes = append(opcoes, otlptracehttp.WithInsecure())
	}

	exportador, err := otlptracehttp.New(ctx, opcoes...)
	if err != nil {
		return nil, fmt.Errorf("criando exportador OTLP: %w", err)
	}

	recurso, err := resource.Merge(
		resource.Default(),
		resource.NewSchemaless(attribute.String("service.name", nomeServico)),
	)
	if err != nil {
		return nil, fmt.Errorf("montando o resource: %w", err)
	}

	provedor := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exportador),
		sdktrace.WithResource(recurso),
	)

	otel.SetTracerProvider(provedor)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return provedor.Shutdown, nil
}

// RegistrarErro marca o span como falho e guarda o erro como evento.
func RegistrarErro(span trace.Span, err error) {
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}
