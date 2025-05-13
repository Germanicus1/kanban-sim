#!/usr/bin/env bash

# === Colors ===
RESET="\033[0m"
RED="\033[0;31m"
GREEN="\033[0;32m"
YELLOW="\033[0;33m"
BLUE="\033[0;34m"

HOST=$1
PORT=$2
shift 2
CMD=$@

echo -e "${BLUE}Waiting for $HOST:$PORT to be ready...${RESET}"

while ! nc -z "$HOST" "$PORT"; do
  sleep 1
done

echo -e "${GREEN}$HOST:$PORT is open, checking Postgres readiness...${RESET}"

while ! docker compose exec -T db psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c '\q' 2>&1; do
  echo -e "${YELLOW}Postgres is not ready yet, retrying...${RESET}"
  sleep 1
done

echo -e "${GREEN}Postgres is ready - executing command${RESET}"
exec $CMD

# make it executabel with: