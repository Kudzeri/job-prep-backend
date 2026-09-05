#!/usr/bin/env sh

set -eu

if [ ! -f "docker-compose.yml" ]; then
	echo "docker-compose.yml not found in current directory"
	exit 1
fi

echo "Starting PostgreSQL via Docker Compose..."
docker compose up -d postgres

echo "Starting API server..."
go run ./cmd/api