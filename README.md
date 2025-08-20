# Witchcraft

Este repositório contém o projeto Witchcraft em Go, que pode ser executado em modo **servidor** ou **cliente** via Docker.

## Como rodar com Docker

### 1. Build da imagem

```bash
docker build -t witchcraft .
```

### 2. Criar Rede Docker (apenas uma vez)
```bash
docker network create witchcraft-net
```

### 3. Rodar Servidor
```bash
docker run -d --name witchcraft-server --network witchcraft-net -p 8080:8080 -e MODE=server witchcraft
```
- -d -> Roda container em segundo plano
- name -> Define nome do container 
- network -> Define qual a rede docker será utilizada
- -p -> configura porta do container para maquina host
- -e -> Define variaveis de ambiente

### 4. Rodar CLiente
```bash
docker run -it --name witchcraft-client --network witchcraft-net -e MODE=client -e SERVER_ADDR=witchcraft-server:8080 witchcraft
```
### 5. Parar e remover containers
```bash
docker rm -f witchcraft-server witchcraft-client
```