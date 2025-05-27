# Kanban Simulation (Work in Progress)

A real-time multiplayer simulation of the getKanban board game, built with Go, PostgreSQL, and Docker.

This project is under active development and currently focuses on backend functionality. The frontend is not yet implemented.

---

## Current Features & Progress

### Backend API

- **Board Configuration**:
  - Configurable board setup via `board_config.json` (columns, subcolumns, WIP limits).
  - Initial card setup based on getKanban v5 configuration.
- **Core Gameplay Logic**:
  - Real-time card movement with support for forward and backward transitions.
  - Day counter with tracking and next day functionality.
  - Deploy logic with lead time calculation (Deploy Day - Selected Day).
  - Card attributes include:
    - Class of Service (Standard, Expedite, Fixed Date, Intangible).
    - Effort estimates for Analysis, Development, and Testing.
    - Value estimate (Low, Medium, High, Very High).
    - Selected Day, Deployed Day, and Lead Time.
- **Game Management**:
  - Game creation and joining by multiple players.
  - Game reset, end, and replay support.
  - Automatic player removal on game end.
  - Player persistence using UUIDs.
- **Database Integration**:
  - PostgreSQL database setup using Docker.
  - Migrations managed with `pressly/goose`.
  - Automated migration execution during `make db-start`.

---

## Setup & Installation

### Prerequisites

- Go (latest version)
- Docker and Docker Compose
- PostgreSQL

### Steps

1. Clone the repository:

   ```sh
   git clone https://github.com/Germanicus1/kanban-sim
   cd backend
   ```

2. Start the database and apply migrations:

   ```sh
   make db-start
   ```

3. Run the backend server:
   ```sh
   go run cmd/main.go
   ```

---

## Makefile Commands

- `make db-new name=<name>` – Create a new migration file.
- `make db-up` – Apply all migrations.
- `make db-down` – Rollback the last migration.
- `make db-status` – Show migration status.
- `make db-reset` – Reset the database (rollback + reapply migrations).
- `make db-start` – Start the database and apply migrations.
- `make db-stop` – Stop the Docker containers.
- `make db-shell` – Open the PostgreSQL shell.

---

## Architecture Pattern:

**Uses a clean architecture with**:

- Repository layer (games.NewSQLRepo)
- Service layer (games.NewService)
- Handler layer (handlers.NewGameHandler, handlers.NewAppHandler)

The server runs on port 8080 and follows REST conventions for the game management API.

## Todo & Upcoming Features

- Implement frontend using ASTRO framework.
- Add real-time multiplayer support via WebSockets.
- Enhance API documentation with OpenAPI specs.

---

## Tech Stack

- **Backend**: Go
- **Database**: PostgreSQL
- **Containerization**: Docker
- **Migrations**: `pressly/goose`

---

## License

This project is licensed under the MIT License.
