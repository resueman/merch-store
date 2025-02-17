#!/bin/bash

sleep 2

echo "Starting test migrations..."

goose -dir "./migrations" postgres "host=${POSTGRES_HOST} port=${POSTGRES_PORT} user=${POSTGRES_USER} password=${POSTGRES_PASSWORD} dbname=${POSTGRES_DB} sslmode=disable" up -v

echo "Test migrations completed."
