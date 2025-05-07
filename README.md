# Kanban Simulation (Work in Progress)

A real-time multiplayer simulation of the getKanban board game, built with Go, PostgreSQL, Docker, and a lightweight ASTRO frontend.

This project is under active development and not yet released.

---

## Current Features & Progress

### Core Gameplay

- Board setup with columns and initial cards based on getKanban v5 configuration
- Real-time card movement, including forward and backward transitions
- Day counter with next day button and day tracking
- Deploy logic with lead time calculation (Deploy Day - Selected Day)
- Configurable board setup via `board_config.json` (columns, subcolumns, WIP limits)
- Card attributes include:

  - Class of Service (Standard, Expedite, Fixed Date, Intangible)
  - Effort estimates for Analysis, Development, and Testing
  - Value estimate (Low, Medium, High, Very High)
  - Selected Day, Deployed Day, and Lead Time

### Game Management

- Game creation and joining by multiple players
- Game reset, end, and replay support
- Automatic player removal on game end
- Player persistence using UUIDs

### Database & Migrations

- Database setup using PostgreSQL with Docker
- Migrations managed with `pressly/goose`
- Automated migration execution during `make db-start`

---

## Setup & Installation

1. **Clone the repository:**

   ```bash
   git clone https://github.com/Germanicus1/kanban-sim.git
   cd kanban-sim
   ```

2. **Environment Variables:**

   Create a `.env` file with the following variables:

   ```
   POSTGRES_USER=germanicus
   POSTGRES_PASSWORD=yourpassword
   POSTGRES_DB=kanbansim
   DB_PORT=5433
   ADMINER_PORT=3334
   ```

3. **Start the Database:**

   ```bash
   make db-start
   ```

4. **Apply Migrations:**

   ```bash
   make db-up
   ```

5. **Run the Server:**

   ```bash
   go run main.go
   ```

---

## Makefile Commands

- `make db-new name=<name>` – Create a new migration file
- `make db-up` – Apply all migrations
- `make db-down` – Rollback the last migration
- `make db-status` – Show migration status
- `make db-reset` – Reset the database (rollback + reapply migrations)
- `make db-start` – Start the database and apply migrations
- `make db-stop` – Stop the Docker containers
- `make db-shell` – Open the PostgreSQL shell

---

## Board Configuration (`board_config.json`)

The game board and initial card setup can be configured via `board_config.json`. Example structure:

```json
{
	"columns": [
		{ "id": "options", "name": "Options" },
		{ "id": "selected", "name": "Selected", "wipLimit": 5 },
		{
			"id": "analysis",
			"name": "Analysis",
			"wipLimit": 3,
			"subcolumns": [
				{ "id": "analysis_in_progress", "name": "In Progress" },
				{ "id": "analysis_done", "name": "Done" }
			]
		}
	],
	"initialCards": [
		{
			"id": "S1",
			"classOfService": "S",
			"columnId": "ready_to_deploy",
			"valueEstimate": "medium",
			"effort": { "analysis": 4, "development": 7, "test": 3 },
			"selectedDay": 1,
			"deployedDay": null
		}
	]
}
```

---

## Todo & Upcoming Features

- Implement additional game rules (expedite lane, blockers, etc.)
- Detailed player management and scoring
- Advanced metrics (CFDs, throughput analysis)
- Integration with a frontend for visual board interaction
- Testing suite with comprehensive test coverage

---

## Tech Stack

- Backend: Go
- Database: PostgreSQL with Docker
- Migrations: `pressly/goose`
- Frontend: ASTRO (upcoming)

---

## License

This project is currently private and not yet released.
