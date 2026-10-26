FROM golang:1.26.10-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/servico-a ./cmd/servico-a && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/servico-b ./cmd/servico-b

FROM alpine:3.22

RUN apk add --no-cache ca-certificates

COPY --from=build /out/servico-a /out/servico-b /usr/local/bin/

USER 65534:65534
EXPOSE 8080 8081

CMD ["servico-a"]
