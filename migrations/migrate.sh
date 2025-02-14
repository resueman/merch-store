#!/bin/bash

sleep 2

echo "Starting migrations..."
goose -dir "${MIGRATION_DIR}" postgres "${MIGRATION_DSN}" up -v
echo "Migrations completed."
