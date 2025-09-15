# Witchcraft

Este repositório contém o projeto Witchcraft em Go, que pode ser executado em modo **servidor**, **cliente** ou **teste de estresse** via Docker e Makefile.

---

## Lista de comandos do Makefile

| Comando                 | O que faz                                                      |
|--------------------------|----------------------------------------------------------------|
| `make network`           | Cria a rede Docker para o projeto                               |
| `make build`             | Faz build da imagem Docker do Witchcraft                        |
| `make run-server`        | Roda o servidor Witchcraft em container Docker                 |
| `make run-client`        | Roda o cliente Witchcraft em container Docker                  |
| `make run-stress`        | Roda o teste de estresse do cliente                             |
| `make run-stress_match`  | Roda o teste de estresse focado em partidas                     |
| `make restart`           | Reinicia o servidor após modificação do código                  |
| `make clean`             | Limpa imagens, containers e rede Docker                          |
| `make help`              | Mostra essa lista de comandos                                    |

> 💡 Algumas variáveis de ambiente podem ser definidas no Makefile ou no `docker run`:
> - `MODE` -> Define se é `server`, `client`, `stress` ou `stress_match`
> - `SERVER_ADDR` -> Endereço IP/hostname do servidor para o cliente ou teste de stress
> - Variáveis de stress (`STRESS_CONCURRENCY`, `STRESS_REQUESTS`, `STRESS_TIMEOUT_MS`, `STRESS_RAMP_MS`) podem ser ajustadas conforme necessidade

---

## Ordem para rodar o servidor e testar com clientes reais

1. Obter o endereço IP da máquina que será o servidor.
2. Definir esse IP na variável de ambiente `SERVER_ADDR` `SERVER_ADDR_UDP` do Dockerfile ou via `docker run`.
3. Rodar:
   ```bash
   make build
   make run-server
   ```
4. Para rodar clientes:
   ```bash
   make run-client
   ```

## Ordem para rodar o servidor e testar o stress (Tudo no mesmo computador)
1. Obter o endereço IP da máquina que será o servidor.
2. Definir esse IP na variável de ambiente `SERVER_ADDR` `SERVER_ADDR_UDP` do Dockerfile ou via `docker run`.
3. Rodar:
   ```bash
   make build
   make run-server
   ```
4. Em outro terminal rodar:
   ```bash
   make run-stres (ou make run-stress_match)
   ```

## Ordem para rodar o servidor e testar o stress (Computadores diferentes)
### primeiro pc
1. Obter o endereço IP da máquina que será o servidor.
2. Definir esse IP na variável de ambiente `SERVER_ADDR` `SERVER_ADDR_UDP` do Dockerfile ou via `docker run`.
3. Rodar:
   ```bash
   make build
   make run-server
   ```

### segundo pc
1. Obter o endereço IP da máquina que será o servidor.
2. Definir esse IP na variável de ambiente `SERVER_ADDR` `SERVER_ADDR_UDP` do Dockerfile ou via `docker run`.
3. Rodar:
   ```bash
   make run-stres (ou make run-stress_match)
   ```

### Observações sobre variáveis de ambiente

- `MODE`: `server`, `client`, `stress` ou `stress_match`
- `SERVER_ADDR`: endereço IP ou hostname do servidor
- Variáveis de teste de stress:
  - `STRESS_CONCURRENCY`: número de conexões simultâneas
  - `STRESS_REQUESTS`: número de requisições por conexão
  - `STRESS_TIMEOUT_MS`: timeout em milissegundos
  - `STRESS_RAMP_MS`: tempo de ramp-up em milissegundos

