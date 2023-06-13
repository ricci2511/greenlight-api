include .env

COMPOSE_FILE ?= docker-compose.dev.yml

# ==================================================================================== #
# HELPERS
# ==================================================================================== #

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

## run/api: run cmd/api application
.PHONY: run/api
run/api:
	go run ./cmd/api -db-dsn=${DATABASE_DSN} -smtp-username=${SMTP_USERNAME} -smtp-password=${SMTP_PASSWORD}

## db/migrations/new name=$1: create a new database migration with the given name
.PHONY: db/migrations/new
db/migrations/new:
	@echo 'Creating new migration files with name ${name}...'
	docker run --rm -v ${PWD}/migrations:/migrations migrate/migrate create -ext sql -dir /migrations -seq ${name}

## db/migrations/up: apply all up database migrations
.PHONY: db/migrations/up
db/migrations/up: confirm
	./wait-for-db.sh ${COMPOSE_FILE}
	@echo 'Running up migrations...'
	docker run --rm -v ${PWD}/migrations:/migrations --network greenlight_default migrate/migrate -path ./migrations -database ${DATABASE_DSN} up

## db/migrations/down: apply all down database migrations
.PHONY: db/migrations/down
db/migrations/down: confirm
	./wait-for-db.sh ${COMPOSE_FILE}
	@echo 'Running down migrations...'
	docker run --rm -v ${PWD}/migrations:/migrations --network greenlight_default migrate/migrate -path ./migrations -database ${DATABASE_DSN} down -all

## db/migrations/force version=$1: force a specific database migration version
.PHONY: db/migrations/force
db/migrations/force: confirm
	./wait-for-db.sh ${COMPOSE_FILE}
	@echo 'Forcing migration version ${version}...'
	docker run --rm -v ${PWD}/migrations:/migrations --network greenlight_default migrate/migrate -path ./migrations -database ${DATABASE_DSN} force ${version}

# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #

## audit: tidy dependencies, format, vet and test all code
.PHONY: audit
audit:
	@echo 'Tidying and verifying dependencies...'
	go mod tidy
	go mod verify
	@echo 'Formatting code...'
	go fmt ./...
	@echo 'Vetting code...'
	go vet ./...
	staticcheck ./...
	@echo 'Running tests...'
	go test -race -vet=off ./...

# ==================================================================================== #
# BUILD
# ==================================================================================== #

## build/api: build the cmd/api application
.PHONY: build/api
build/api:
	@echo 'Building cmd/api...'
	go build -ldflags='-s' -o=./bin/api ./cmd/api
	GOOS=linux GOARCH=amd64 go build -ldflags='-s' -o=./bin/linux_amd64/api ./cmd/api
