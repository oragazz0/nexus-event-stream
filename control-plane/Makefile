include .env
export

.PHONY: help db-up db-down install migrate superuser run setup

help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

db-up: ## Start PostgreSQL container
	docker compose -f ../infrastructure/docker-compose.yml up -d

db-down: ## Stop PostgreSQL container
	docker compose -f ../infrastructure/docker-compose.yml down

install: ## Install Python dependencies
	pip install -r requirements.txt

migrate: ## Apply database migrations
	python manage.py migrate

superuser: ## Create superuser from environment variables
	python manage.py createsuperuser --noinput || true

run: ## Run development server
	python manage.py runserver

tests: ## Run tests
	python manage.py test nexus.core.tests.SignalEventPublishingTestCase -v 2

setup: db-up install migrate superuser ## Full setup: database, dependencies, migrations, superuser
	@echo "Setup complete! Run 'make run' to start the server."
