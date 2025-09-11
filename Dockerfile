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
# Mudar aqui e UDP para máquina HOST
ENV MODE=server
ENV SERVER_ADDR=192.168.0.101:8080 
ENV UDP_SERVER_ADDR=192.168.0.101:9999
EXPOSE 8080
EXPOSE 9999

CMD ["./app"]
