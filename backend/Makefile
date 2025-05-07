# === Load .env ===
include .env
export

# === Config ===
DB_URL := postgresql://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:$(DB_PORT)/$(POSTGRES_DB)?sslmode=disable
MIGRATIONS_DIR := ./migrations

# === Commands ===
.PHONY: help db-new db-up db-down db-status db-reset db-start db-stop db-shell

help:
	@echo ""
	@echo "Kanban-Sim Docker DB Commands:"
	@echo "  make db-new name=init_schema   # Create new migration"
	@echo "  make db-up                     # Apply all migrations"
	@echo "  make db-down                   # Rollback last migration"
	@echo "  make db-status                 # Show migration status"
	@echo "  make db-reset                  # Reset database (down + up)"
	@echo "  make db-start                  # Start Docker containers"
	@echo "  make db-stop                   # Stop Docker containers"
	@echo "  make db-shell                  # Open psql shell"
	@echo ""


db-new:
ifndef name
	$(error "❌ Please provide a name: make db-new name=init_schema")
endif
	goose -dir $(MIGRATIONS_DIR) create $(name) sql

db-up:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" up

db-down:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" down

db-status:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" status

db-reset:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" reset

db-start:
	docker compose up -d

db-stop:
	docker compose down

db-shell:
	docker compose exec db psql -U $(POSTGRES_USER) -d $(POSTGRES_DB)
