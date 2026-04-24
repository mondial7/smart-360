#!/bin/bash

# Script to reset and reseed the database with comprehensive development data
# Usage: ./reseed-dev.sh

set -e

# Check if MongoDB is running
if ! docker ps | grep -q mongodb 2>/dev/null; then
    echo "⚠️  MongoDB container is not running. Starting it..."
    docker-compose up -d mongodb
    echo "⏳ Waiting for MongoDB to be ready..."
    sleep 3
    echo ""
fi

# Run the seed program
go run cmd/seed/main.go
