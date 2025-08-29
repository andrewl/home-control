# Project name
APP_NAME = control-panel

# Default target
.PHONY: help
help: ## Show this help
	@echo "Usage:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

.PHONY: build
build: ## Build Docker image
	docker compose build

.PHONY: up
up: ## Run the app (foreground)
	docker compose up

.PHONY: up-d
up-d: ## Run the app (detached)
	docker compose up -d

.PHONY: down
down: ## Stop and remove container
	docker compose down

.PHONY: restart
restart: ## Restart the app
	docker compose restart

.PHONY: logs
logs: ## Tail logs
	docker compose logs -f

.PHONY: clean
clean: ## Remove containers, networks, volumes, and images
	docker compose down --rmi all --volumes --remove-orphans

.PHONY: run-local
run-local: ## Run the Go app locally without Docker
	go run main.go

.PHONY: test
test: ## Run Go unit tests with coverage
	go test ./... -v -cover

