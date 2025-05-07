# Kanban Simulation (Work in Progress)

This is a real-time multiplayer simulation of the getKanban board game. Built
with Go, postgres (docker), and a lightweight ASTRO frontend.

> ⚠️ This project is **not released** yet. It's under active development.

## Features

- Real-time card movement and board updates
- Configurable board via JSON
- Lead time tracking and value delivery simulation
- Reset, end, and replay support

## Setup

1. Clone the repo
2. Set up a Supabase project
3. Create `.env` with `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`, `DB_PORT` and `ADMINER_PORT`
4. Run the server:

```
go run main.go
```

## License

Private / Not yet released.
