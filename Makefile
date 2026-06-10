MANAGED =  provider1 provider2 core

MANAGER = monitor analyze plan execute knowledge

SERVICES = $(MANAGED) $(MANAGER)

.PHONY: help build up down logs clean start-% stop-% logs-% build-%

COMPOSE = docker compose -f docker-compose.yml

help:
	@echo "Comandos disponíveis:"
	@echo "  make build
	@echo "  make up
	@echo "  make down
	@echo "  make logs
	@echo ""
	@echo "Serviços individuais:"
	@echo "  make start-<service>"
	@echo "  make stop-<service>"
	@echo "  make logs-<service>"

build:
	$(COMPOSE) build

up:
	$(COMPOSE) up -d

down:
	$(COMPOSE) down

logs:
	$(COMPOSE) logs -f

clean:
	$(COMPOSE) down -v --remove-orphans

start-manager:
	$(COMPOSE) up -d $(MANAGER)

stop-manager:
	$(COMPOSE) stop $(MANAGER)

start-managed:
	$(COMPOSE) up -d $(MANAGED)

stop-managed:
	$(COMPOSE) stop $(MANAGED)

start-%:
	$(COMPOSE) up -d $*

stop-%:
	$(COMPOSE) stop $*

logs-%:
	$(COMPOSE) logs -f $*

build-%:
	$(COMPOSE) build $*