# === Load .env ===
include .env
export

# === Config ===
DB_URL := postgresql://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:$(DB_PORT)/$(POSTGRES_DB)?sslmode=disable
MIGRATIONS_DIR := ./migrations

# === Colors ===
RESET := \033[0m
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
BLUE := \033[0;34m

# === Commands ===
.PHONY: help db-new db-up db-down db-status db-reset db-start db-stop db-shell

help:
	@echo ""
	@echo "$(YELLOW)Kanban-Sim Docker DB Commands:$(RESET)"
	@echo "  $(BLUE)make db-new name=init_schema$(RESET)   # Create new migration"
	@echo "  $(BLUE)make db-up$(RESET)                     # Apply all migrations"
	@echo "  $(BLUE)make db-down$(RESET)                   # Rollback last migration"
	@echo "  $(BLUE)make db-status$(RESET)                 # Show migration status"
	@echo "  $(BLUE)make db-reset$(RESET)                  # Reset database (down + up)"
	@echo "  $(BLUE)make db-start$(RESET)                  # Start Docker containers"
	@echo "  $(BLUE)make db-stop$(RESET)                   # Stop Docker containers"
	@echo "  $(BLUE)make db-shell$(RESET)                  # Open psql shell"
	@echo ""

db-new:
ifndef name
	$(error "$(RED)❌ Please provide a name: make db-new name=init_schema$(RESET)")
endif
	@echo "$(GREEN)Creating new migration: $(name)...$(RESET)"
	goose -dir $(MIGRATIONS_DIR) create $(name) sql

db-up:
	@echo "$(GREEN)Applying all migrations...$(RESET)"
	goose -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" up

db-down:
	@echo "$(YELLOW)Rolling back last migration...$(RESET)"
	goose -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" down

db-status:
	@echo "$(BLUE)Checking migration status...$(RESET)"
	goose -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" status

db-reset:
	@echo "$(RED)Resetting database...$(RESET)"
	goose -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" reset
	@echo "$(GREEN)Reapplying migrations...$(RESET)"
	$(MAKE) db-up

db-start:
	@echo "$(GREEN)Starting Docker containers...$(RESET)"
	docker compose up -d

	@echo "$(BLUE)Waiting for database to be ready...$(RESET)"
	./wait-for-it.sh localhost $(DB_PORT) -- $(MAKE) db-up

db-stop:
	@echo "$(YELLOW)Stopping Docker containers...$(RESET)"
	docker compose down

db-shell:
	@echo "$(BLUE)Opening psql shell...$(RESET)"
	docker compose exec db psql -U $(POSTGRES_USER) -d $(POSTGRES_DB)
