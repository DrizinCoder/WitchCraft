# Etapa 1: build
FROM golang:1.24.5 AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN go build -o app .

# Etapa 2: execução
FROM golang:1.24.5

WORKDIR /root/

COPY --from=builder /app/app .

# Usa variável de ambiente MODE
ENV MODE=server
ENV SERVER_ADDR=127.0.0.1:8080
ENV UDP_SERVER_ADDR=server:9999

CMD ["./app"]
