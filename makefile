.PHONY: help up down test run clean

help:
	@echo "📈 Market Aggregator - Command Menu"
	@echo "-----------------------------------"
	@echo "make up      		- Start the full Docker infrastructure (DB + Engine)"
	@echo "make down    		- Stop and remove the Docker containers and volumes"
	@echo "make test    		- Run all Go tests with the race detector"
	@echo "make run     		- Run the Go engine locally (requires local DB)"
	@echo "make clean   		- Remove any compiled binaries"
	@echo "make infra-init	- runs tofu init command"
	@echo "make infra-plan	- runs tofu plan command"
	@echo "make infra-apply	- runs tofu apply command"
	@echo "make infra-destroy	- runs tofu destroy command"

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

infra-init:
	cd infra && tofu init

infra-plan:
	cd infra && tofu plan -var="db_password=SuperSecret123!"

infra-apply:
	cd infra && tofu apply -var="db_password=SuperSecret123!"
	
infra-destroy:
	cd infra && tofu destroy