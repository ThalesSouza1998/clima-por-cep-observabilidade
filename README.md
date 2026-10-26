# Clima por CEP com Observabilidade

Sistema distribuído em Go com dois microsserviços, OpenTelemetry e Zipkin. O
Serviço A recebe o CEP, valida e encaminha para o Serviço B. O Serviço B
consulta o ViaCEP, busca a temperatura na WeatherAPI e devolve a cidade com as
temperaturas em Celsius, Fahrenheit e Kelvin.

## Arquitetura

```
cliente
  │  POST / {"cep":"29902555"}
  ▼
serviço A (porta 8080)
  │  POST / {"cep":"29902555"}  (propaga o traceparent)
  ▼
serviço B (porta 8081)
  ├── HTTP GET → ViaCEP      (span manual: busca-cep)
  └── HTTP GET → WeatherAPI  (span manual: busca-temperatura)

serviço A ─┐
           ├── OTLP HTTP (4318) → OTEL Collector → Zipkin (9411)
serviço B ─┘
```

## Como rodar

Pré-requisitos: Docker e Docker Compose.

```bash
cp .env.example .env
# edite o .env e informe a chave da WeatherAPI em WEATHER_API_KEY

docker compose up --build
```

O Compose sobe os quatro componentes:

| Componente    | Endereço local          |
| ------------- | ----------------------- |
| Serviço A     | http://localhost:8080   |
| Serviço B     | interno, `servico-b:8081` |
| OTEL Collector | portas 4317 e 4318     |
| Zipkin        | http://localhost:9411   |

O Serviço B não é publicado no host: o fluxo esperado é sempre passar pelo
Serviço A.

## Fazendo uma requisição

```bash
curl -i -X POST http://localhost:8080/ \
  -H 'Content-Type: application/json' \
  -d '{"cep":"29902555"}'
```

Resposta:

```json
{
  "city": "Linhares",
  "temp_C": 23,
  "temp_F": 73.4,
  "temp_K": 296
}
```

As temperaturas mudam conforme o momento da consulta.

## Erros

| Status | Corpo                                   | Quando acontece |
| ------ | --------------------------------------- | --------------- |
| 422    | `{"error":"invalid zipcode"}`           | O CEP não é string ou não tem exatamente 8 dígitos |
| 404    | `{"error":"can not find zipcode"}`       | O CEP tem formato válido, mas não existe no ViaCEP |
| 500    | `{"error":"internal error"}`             | Falha inesperada ao consultar as APIs externas |
| 502    | `{"error":"service unavailable"}`        | O Serviço A não conseguiu falar com o Serviço B |
| 405    | —                                       | Método diferente de POST |

O Serviço A valida o CEP antes de encaminhar e o Serviço B valida novamente,
para que ele também possa ser chamado diretamente durante o desenvolvimento.

## Observabilidade

Os dois serviços exportam traces via OTLP HTTP para o OTEL Collector. O
Collector processa os lotes e envia para o Zipkin.

- As requisições HTTP de entrada são instrumentadas com `otelhttp`.
- A chamada do Serviço A para o Serviço B também é instrumentada, e o
  `traceparent` é propagado automaticamente.
- O Serviço B cria spans manuais em volta das chamadas externas:
  - `busca-cep`, com os atributos `cep` e `cidade`;
  - `busca-temperatura`, com os atributos `cidade` e
    `temperatura_celsius`.

O trace completo fica assim:

```
servico-a  POST /                    834 ms
└─ servico-a  HTTP POST              834 ms
   └─ servico-b  POST /              832 ms
      ├─ servico-b  busca-cep        430 ms
      └─ servico-b  busca-temperatura 402 ms
```

Os tempos variam a cada execução; o importante é a hierarquia dos spans.

### Vendo no Zipkin

1. Acesse http://localhost:9411.
2. Selecione o serviço `servico-a` e clique em **Run Query**.
3. Abra o trace encontrado.

Os spans são enviados em lote, então pode levar alguns segundos até aparecerem.
Também é possível consultar pela API:

```bash
curl http://localhost:9411/api/v2/services
```

## Configuração

| Variável | Usada por | Padrão | Descrição |
| -------- | --------- | ------ | --------- |
| `WEATHER_API_KEY` | Serviço B | — | Chave da WeatherAPI (obrigatória) |
| `SERVICO_B_URL` | Serviço A | `http://localhost:8081` | URL do Serviço B |
| `PORT` | A e B | `8080` / `8081` | Porta HTTP |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | A e B | `http://localhost:4318` | Endpoint OTLP HTTP do Collector |

No Compose, o endpoint do Collector e a URL do Serviço B já são configurados
pelas variáveis de ambiente.

## Rodando sem Docker

Para desenvolver localmente, rode os serviços em terminais separados:

```bash
go run ./cmd/servico-b
go run ./cmd/servico-a
```

Nesse caso o Serviço B lê a `WEATHER_API_KEY` do `.env`. Sem um Collector em
`localhost:4318`, a API continua funcionando, mas o SDK registra falhas de
exportação de traces no log.

## Testes

```bash
go test ./...
```

Os testes cobrem validação de CEP, conversões, regras do Serviço B, respostas
HTTP dos dois serviços, criação dos spans manuais e propagação do
`traceparent` entre A e B.

## Estrutura

```
cmd/
  servico-a/          entrada HTTP
  servico-b/          orquestração
internal/
  cep/                validação de CEP
  clima/              regras e conversões
  viacep/             cliente ViaCEP + span manual
  weatherapi/         cliente WeatherAPI + span manual
  servicoa/           handler do Serviço A
  servicob/           handler do Serviço B
  telemetria/         configuração do OpenTelemetry
  httpjson/           respostas JSON
otel-collector-config.yaml
docker-compose.yaml
Dockerfile
```

## Detalhes de implementação

- A conversão para Kelvin usa `273`. O Fahrenheit é arredondado para uma casa
  decimal e o Kelvin para duas.
- As chamadas externas têm timeout de 10 segundos; a chamada de A para B tem
  timeout de 15 segundos.
- O `service.name` do resource é `servico-a` ou `servico-b`, e é o nome que
  aparece no Zipkin.
- O `.env` não é versionado nem entra na imagem Docker.
