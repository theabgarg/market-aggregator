.PHONY: help up down test run clean

help:
	@echo "📈 Market Aggregator - Command Menu"
	@echo "-----------------------------------"
	@echo "make up      - Start the full Docker infrastructure (DB + Engine)"
	@echo "make down    - Stop and remove the Docker containers and volumes"
	@echo "make test    - Run all Go tests with the race detector"
	@echo "make run     - Run the Go engine locally (requires local DB)"
	@echo "make clean   - Remove any compiled binaries"

up:
	docker-compose up --build

down:
	docker-compose down -v

test:
	go test -race ./...

run:
	go run cmd/aggregator/main.go

clean:
	rm -f aggregator