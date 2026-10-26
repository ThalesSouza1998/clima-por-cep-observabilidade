package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/ThalesSouza1998/clima-por-cep-observabilidade/internal/clima"
	"github.com/ThalesSouza1998/clima-por-cep-observabilidade/internal/servicob"
	"github.com/ThalesSouza1998/clima-por-cep-observabilidade/internal/telemetria"
	"github.com/ThalesSouza1998/clima-por-cep-observabilidade/internal/viacep"
	"github.com/ThalesSouza1998/clima-por-cep-observabilidade/internal/weatherapi"
)

func main() {
	_ = godotenv.Load()

	chave := os.Getenv("WEATHER_API_KEY")
	if chave == "" {
		log.Fatal("WEATHER_API_KEY não definida")
	}

	encerrar, err := telemetria.Iniciar(context.Background(), "servico-b")
	if err != nil {
		log.Fatalf("iniciando a telemetria: %v", err)
	}
	defer encerrarTelemetria(encerrar)

	servico := clima.NovoServico(viacep.Novo(), weatherapi.Novo(chave))

	porta := os.Getenv("PORT")
	if porta == "" {
		porta = "8081"
	}

	handler := servicob.Novo(servico)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /", handler.Clima)

	log.Printf("serviço B escutando na porta %s", porta)
	if err := http.ListenAndServe(":"+porta, otelhttp.NewHandler(mux, "servico-b")); err != nil {
		log.Fatal(err)
	}
}

func encerrarTelemetria(encerrar func(context.Context) error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := encerrar(ctx); err != nil {
		log.Printf("encerrando a telemetria: %v", err)
	}
}
