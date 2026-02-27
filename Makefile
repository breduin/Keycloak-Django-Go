NETWORK_NAME := sso-example-net

.PHONY: network init-envs build-all up-all down-all up-keycloak down-keycloak up-django down-django up-go down-go logs first-run

network:
	docker network create $(NETWORK_NAME) 2>/dev/null || true

init-envs:
	test -f django_app/.env || cp django_app/.env.example django_app/.env
	test -f go_app/.env || cp go_app/.env.example go_app/.env

build-all:
	docker compose -f keycloak/docker-compose.yml build
	docker compose -f django_app/docker-compose.yml build
	docker compose -f go_app/docker-compose.yml build

up-keycloak: network
	docker compose -f keycloak/docker-compose.yml up -d

down-keycloak:
	docker compose -f keycloak/docker-compose.yml down

up-django: network
	docker compose -f django_app/docker-compose.yml up -d

down-django:
	docker compose -f django_app/docker-compose.yml down

up-go: network
	docker compose -f go_app/docker-compose.yml up -d

down-go:
	docker compose -f go_app/docker-compose.yml down

up-all: up-keycloak up-django up-go

down-all:
	docker compose -f keycloak/docker-compose.yml down
	docker compose -f django_app/docker-compose.yml down
	docker compose -f go_app/docker-compose.yml down

logs:
	docker compose -f keycloak/docker-compose.yml logs -f

first-run: network init-envs build-all up-all
	@echo "Первый запуск завершён. Keycloak: http://localhost:8210, Django: http://localhost:8211, Go: http://localhost:8212"

