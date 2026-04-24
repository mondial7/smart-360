.PHONY: help docker-up docker-down docker-logs dev build clean

# Default target
help:
	@echo "Smart 360 Development Commands:"
	@echo ""
	@echo "  docker-up     - Start MongoDB with Docker"
	@echo "  docker-down   - Stop MongoDB containers"
	@echo "  docker-logs   - Show MongoDB logs"
	@echo "  dev           - Start development server"
	@echo "  build         - Build the application"
	@echo "  clean         - Clean Docker containers and volumes"
	@echo ""

# Docker commands
docker-up:
	@echo "Starting MongoDB..."
	docker-compose up -d mongodb
	@echo "MongoDB started at mongodb://localhost:27017"
	@echo "Mongo Express available at http://localhost:8081"

docker-down:
	@echo "Stopping MongoDB..."
	docker-compose down

docker-logs:
	@echo "Showing MongoDB logs..."
	docker-compose logs -f mongodb

# Development commands
dev: docker-up
	@echo "Starting development server..."
	@echo "Make sure to copy .env.example to .env and configure your settings"
	go run main.go

# Build commands
build:
	@echo "Building application..."
	go build -o smart360-server main.go

# Cleanup commands
clean:
	@echo "Cleaning Docker containers and volumes..."
	docker-compose down -v
	docker system prune -f

# Quick start for new developers
setup:
	@echo "Setting up Smart 360 development environment..."
	@echo ""
	@echo "1. Starting MongoDB..."
	docker-compose up -d mongodb
	@echo ""
	@echo "2. Copying environment configuration..."
	cp .env.example .env
	@echo ""
	@echo "3. Installing Go dependencies..."
	go mod tidy
	@echo ""
	@echo "Setup complete!"
	@echo ""
	@echo "Next steps:"
	@echo "1. Edit .env with your Google OAuth credentials"
	@echo "2. Run 'make dev' to start the development server"
	@echo "3. Visit http://localhost:8081 for Mongo Express (admin/admin123)"
	@echo "4. Start your frontend application"
