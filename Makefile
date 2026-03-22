.PHONY: run-api build-api run-frontend build-frontend install-frontend migrate-up migrate-down migrate-up-docker migrate-down-docker db-shell make-script-exec simulate-telemetry

MIGRATIONS_DIR := migrations
POSTGRES_SERVICE := postgres
POSTGRES_USER := solar_user
POSTGRES_DB := solar_db
BUILD_GOOS := linux
BUILD_GOARCH := amd64

run-api:
	go run ./cmd/api

build-api:
	CGO_ENABLED=0 GOOS=$(BUILD_GOOS) GOARCH=$(BUILD_GOARCH) go build -o solar-be ./cmd/api

migrate-up:
	@set -a; . ./.env; set +a; \
	for file in $(MIGRATIONS_DIR)/*.up.sql; do \
		echo "Applying $$file"; \
		psql "$$DATABASE_URL" -v ON_ERROR_STOP=1 -f "$$file"; \
	done

migrate-down:
	@set -a; . ./.env; set +a; \
	for file in $$(ls $(MIGRATIONS_DIR)/*.down.sql | sort -r); do \
		echo "Reverting $$file"; \
		psql "$$DATABASE_URL" -v ON_ERROR_STOP=1 -f "$$file"; \
	done

migrate-up-docker:
	@for file in $(MIGRATIONS_DIR)/*.up.sql; do \
		echo "Applying $$file to Docker PostgreSQL"; \
		docker compose exec -T $(POSTGRES_SERVICE) psql -U $(POSTGRES_USER) -d $(POSTGRES_DB) -v ON_ERROR_STOP=1 -f - < "$$file"; \
	done

migrate-down-docker:
	@for file in $$(ls $(MIGRATIONS_DIR)/*.down.sql | sort -r); do \
		echo "Reverting $$file from Docker PostgreSQL"; \
		docker compose exec -T $(POSTGRES_SERVICE) psql -U $(POSTGRES_USER) -d $(POSTGRES_DB) -v ON_ERROR_STOP=1 -f - < "$$file"; \
	done

db-shell:
	docker compose exec $(POSTGRES_SERVICE) psql -U $(POSTGRES_USER) -d $(POSTGRES_DB)

run-frontend:
	cd frontend && npm run dev

build-frontend:
	cd frontend && npm run build

install-frontend:
	cd frontend && npm install

make-script-exec:
	chmod +x requests/simulate_telemetry.sh

simulate-telemetry: make-script-exec
	@echo "Usage: DEVICE_KEY=dvc_xxx DEVICE_ID=plant-A-01 make simulate-telemetry"
	@BASE_URL=$${BASE_URL:-http://localhost:8080} DEVICE_KEY=$$DEVICE_KEY DEVICE_ID=$${DEVICE_ID:-plant-A-01} POINTS=$${POINTS:-6} INTERVAL_MINUTES=$${INTERVAL_MINUTES:-10} requests/simulate_telemetry.sh
