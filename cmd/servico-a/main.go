package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/ThalesSouza1998/clima-por-cep-observabilidade/internal/servicoa"
	"github.com/ThalesSouza1998/clima-por-cep-observabilidade/internal/telemetria"
)

func main() {
	_ = godotenv.Load()

	encerrar, err := telemetria.Iniciar(context.Background(), "servico-a")
	if err != nil {
		log.Fatalf("iniciando a telemetria: %v", err)
	}
	defer encerrarTelemetria(encerrar)

	destino := os.Getenv("SERVICO_B_URL")
	if destino == "" {
		destino = "http://localhost:8081"
	}

	porta := os.Getenv("PORT")
	if porta == "" {
		porta = "8080"
	}

	handler := servicoa.Novo(destino)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /", handler.Clima)

	log.Printf("serviço A escutando na porta %s (serviço B: %s)", porta, destino)
	if err := http.ListenAndServe(":"+porta, otelhttp.NewHandler(mux, "servico-a")); err != nil {
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
