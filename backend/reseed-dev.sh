#!/bin/bash

# Script to reset and reseed the database with comprehensive development data
# Usage: ./reseed-dev.sh   (runs from any cwd — script switches to its own dir)

set -e

cd "$(dirname "$0")"

if ! docker ps --format '{{.Names}}' | grep -q '^smart360-mongodb$'; then
    echo "⚠️  MongoDB container is not running. Starting it..."
    docker-compose up -d mongodb
    echo "⏳ Waiting for MongoDB to be ready..."
    sleep 3
    echo ""
fi

go run cmd/seed/main.go
