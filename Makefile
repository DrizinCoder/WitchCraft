# Variáveis para o nome do projeto, imagem e container
PROJECT_NAME = witchcraft
IMAGE_NAME = $(PROJECT_NAME)
SERVER_CONTAINER = $(PROJECT_NAME)-server
NETWORK_NAME = $(PROJECT_NAME)-net
CLIENT_IMAGE = $(PROJECT_NAME)

# --- Alvos principais ---

.PHONY: network
network:
	@echo "Criando a rede Docker '$(NETWORK_NAME)'..."
	-docker network create $(NETWORK_NAME)

.PHONY: build
build:
	@echo "Construindo a imagem '$(IMAGE_NAME)'..."
	docker build -t $(IMAGE_NAME) .

.PHONY: run-server
run-server:
	@echo "Rodando o container do servidor '$(SERVER_CONTAINER)'..."
	docker run --name $(SERVER_CONTAINER) --network $(NETWORK_NAME) -p 8080:8080 -e MODE=server $(IMAGE_NAME)

.PHONY: run-client
run-client:
	@echo "Rodando o container do cliente..."
	docker run -it --network $(NETWORK_NAME) -e MODE=client -e SERVER_ADDR=$(SERVER_CONTAINER):8080 $(CLIENT_IMAGE)

.PHONY: rm-server
rm-server:
	@echo "Removendo o container do servidor '$(SERVER_CONTAINER)'..."
	-docker rm -f $(SERVER_CONTAINER) || true

.PHONY: restart
restart: rm-server build run-server
	@echo "Servidor reiniciado com sucesso!"

.PHONY: clean
clean: rm-server
	@echo "Removendo a imagem '$(IMAGE_NAME)'..."
	-docker rmi $(IMAGE_NAME) || true
	@echo "Removendo a rede '$(NETWORK_NAME)'..."
	-docker network rm $(NETWORK_NAME) || true
	@echo "Limpeza completa!"

.PHONY: help
help:
	@echo "--- Lista de Comandos Makefile ---"
	@echo "network -> Cria Rede docker"
	@echo "build -> Builda Imagem"
	@echo "run-server -> Roda container do servidor"
	@echo "run-client -> Roda container do cliente"
	@echo "restart -> Reinicia server após modificação de código"
	@echo "clean -> Limpa projeto"

.PHONY: run-stress
run-stress:
	@echo "Rodando stress test contra '$(SERVER_CONTAINER):8080'..."
	docker run --rm --network $(NETWORK_NAME) \
		-e MODE=stress \
		-e SERVER_ADDR=$(SERVER_CONTAINER):8080 \
		-e STRESS_CONCURRENCY=21000 \
		-e STRESS_REQUESTS=5 \
		-e STRESS_TIMEOUT_MS=2000 \
		-e STRESS_RAMP_MS=0 \
		$(CLIENT_IMAGE)