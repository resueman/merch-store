#!/bin/bash

export MIGRATION_DIR=./migrations
export MIGRATION_DSN="host=$POSTGRES_HOST port=$POSTGRES_PORT dbname=$POSTGRES_DB user=$POSTGRES_USER password=$POSTGRES_PASSWORD sslmode=disable"

sleep 2

echo "Starting migrations..."
goose -dir "${MIGRATION_DIR}" postgres "${MIGRATION_DSN}" up -v
echo "Migrations completed."
