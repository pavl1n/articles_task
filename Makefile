run:
	@set -a; [ -f .env ] && . ./.env; set +a; go run ./cmd/api

test:
	go test ./...

migrate-up:
	@set -a; [ -f .env ] && . ./.env; set +a; migrate -path db/migrations -database "$${DATABASE_URL}" up

migrate-down:
	@set -a; [ -f .env ] && . ./.env; set +a; migrate -path db/migrations -database "$${DATABASE_URL}" down 1

up:
	docker-compose up --build
