default: help

PROJECTNAME=$(shell basename "$(PWD)")
DC := docker compose

DB_DSN = postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)

MIGRATIONS_DIR = db/migrations
COLOR_RESET = \033[0m
COLOR_BOLD = \033[1m
COLOR_GREEN = \033[32m
COLOR_YELLOW = \033[33m
COLOR_RED = \033[31m

ifneq (,$(wildcard ./.env))
    include .env
    export
endif

.PHONY: lint
lint:
	@golangci-lint run ./... --fix

.PHONY: setup
setup:
	@cp -n .env.example .env

# ======== Для локального использования ========
.PHONY: db-create
db-create: ## Создать базу данных
	@EXISTS=$$(PGPASSWORD="$(DB_PASSWORD)" \
		psql -X -tAq \
		  -h "$(DB_HOST)" -p "$(DB_PORT)" -U "$(DB_USER)" -d postgres \
		  -c "SELECT 1 FROM pg_database WHERE datname='$(DB_NAME)';"); \
	if [ "$$EXISTS" = "1" ]; then \
	    echo "$(COLOR_YELLOW)⚠️  Database $(DB_NAME) already exists$(COLOR_RESET)"; \
	else \
	    echo "$(COLOR_YELLOW)Creating database $(DB_NAME)...$(COLOR_RESET)"; \
	    PGPASSWORD="$(DB_PASSWORD)" \
	    psql -X -q \
	      -h "$(DB_HOST)" -p "$(DB_PORT)" -U "$(DB_USER)" -d postgres \
	      -c "CREATE DATABASE \"$(DB_NAME)\";" >/dev/null 2>&1 && \
	    echo "$(COLOR_GREEN)✅ Database $(DB_NAME) created$(COLOR_RESET)"; \
	fi


.PHONY: db-drop
db-drop: ## Удалить базу данных
	@EXISTS=$$(PGPASSWORD="$(DB_PASSWORD)" \
		psql -X -tAq \
		  -h "$(DB_HOST)" -p "$(DB_PORT)" -U "$(DB_USER)" -d postgres \
		  -c "SELECT 1 FROM pg_database WHERE datname='$(DB_NAME)';"); \
	if [ "$$EXISTS" = "1" ]; then \
	    echo "$(COLOR_YELLOW)🗑  Dropping database $(DB_NAME)...$(COLOR_RESET)"; \
	    PGPASSWORD="$(DB_PASSWORD)" \
	    psql -X -q \
	      -h "$(DB_HOST)" -p "$(DB_PORT)" -U "$(DB_USER)" -d postgres \
	      -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname='$(DB_NAME)';" >/dev/null 2>&1 || true; \
	    PGPASSWORD="$(DB_PASSWORD)" \
	    psql -X -q \
	      -h "$(DB_HOST)" -p "$(DB_PORT)" -U "$(DB_USER)" -d postgres \
	      -c "DROP DATABASE \"$(DB_NAME)\";" >/dev/null 2>&1 && \
	    echo "$(COLOR_GREEN)✅ Database $(DB_NAME) dropped$(COLOR_RESET)"; \
	else \
	    echo "$(COLOR_RED)❌ Database $(DB_NAME) does not exist$(COLOR_RESET)"; \
	fi

.PHONY: run
run:
	@echo "$(COLOR_YELLOW)Starting application...$(COLOR_RESET)"
	@go run ./cmd/api/main.go

# ====== Docker development ======

.PHONY: db-migrate
#	@goose -dir $(MIGRATIONS_DIR) postgres "$(DB_DSN)" up"
db-migrate:
	$(DC) run --rm migrate up

.PHONY: db-rollback
# 	@goose -dir $(MIGRATIONS_DIR) postgres "$(DB_DSN)" down
db-rollback:
	$(DC) run --rm migrate down

.PHONY: migrate-status
# 	@goose -dir $(MIGRATIONS_DIR) postgres "$(DB_DSN)" status
migrate-status:
	$(DC) run --rm migrate status

# Запустить все сервисы
up:
	$(DC) up

# Пересобрать и запустить
up-build:
	$(DC) up --build

# Остановить и убрать контейнеры (TODO: --remove-orphans)
down:
	$(DC) down

# Остановить + удалить volumes (полный вайп)
down-hard:
	$(DC) down --rmi all --volumes --remove-orphans

# Перезапустить контейнеры
restart:
	$(DC) down
	$(DC) up -d

shell:
	$(DC) run --rm app bash

db-shell:
	$(DC) exec db psql -U $(DB_USER) -d $(DB_NAME)

redis-shell:
	$(DC) exec redis redis-cli

# Показать статус контейнеров
health:
	$(DC) ps --format "table {{.Name}}\t{{.Status}}"

logs:
	$(DC) logs -f

logs-app:
	$(DC) logs -f app

logs-db:
	$(DC) logs -f db

logs-redis:
	$(DC) logs -f redis

.PHONY: help
help:
	@echo "$(COLOR_BOLD)Available commands:$(COLOR_RESET)"
	@echo "  $(COLOR_GREEN)up$(COLOR_RESET)              - Start all services"
	@echo "  $(COLOR_GREEN)up-build$(COLOR_RESET)        - Rebuild and start"
	@echo "  $(COLOR_GREEN)down$(COLOR_RESET)            - Stop services"
	@echo "  $(COLOR_GREEN)down-hard$(COLOR_RESET)       - Stop + remove everything"
	@echo "  $(COLOR_GREEN)restart$(COLOR_RESET)         - Restart all services"
	@echo ""
	@echo "  $(COLOR_YELLOW)db-create$(COLOR_RESET)       - Create database"
	@echo "  $(COLOR_YELLOW)db-drop$(COLOR_RESET)         - Drop database"
	@echo "  $(COLOR_YELLOW)db-migrate$(COLOR_RESET)      - Run migrations"
	@echo "  $(COLOR_YELLOW)db-rollback$(COLOR_RESET)     - Rollback last migration"
	@echo "  $(COLOR_YELLOW)migrate-status$(COLOR_RESET)  - Show migration status"
	@echo ""
	@echo "  $(COLOR_GREEN)shell$(COLOR_RESET)           - Open shell in app container"
	@echo "  $(COLOR_GREEN)db-shell$(COLOR_RESET)        - Open PostgreSQL shell"
	@echo "  $(COLOR_GREEN)redis-shell$(COLOR_RESET)     - Open Redis CLI"
	@echo ""
	@echo "  $(COLOR_GREEN)logs$(COLOR_RESET)            - Show all logs"
	@echo "  $(COLOR_GREEN)health$(COLOR_RESET)          - Show container status"