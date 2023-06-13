#!/bin/bash

COMPOSE_FILE=$1

# Helper function to wait for the postgres database to start.
wait_for_db() {
  while ! pg_isready -U "${DATABASE_USER}" -d "${DATABASE_NAME}" -h localhost; do
    sleep 1
  done
}

# Load environment variables from .env file to OS environment.
if [[ -f .env ]]; then
  export $(grep -v '^#' .env | xargs)
fi

# Start database container from the compose file.
docker-compose -f "$COMPOSE_FILE" up -d db

echo 'Waiting for database to start...'

# Run the wait_for_db helper in the db container.
docker-compose -f "$COMPOSE_FILE" exec db /bin/sh -c "$(declare -f wait_for_db); wait_for_db"
