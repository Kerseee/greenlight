## include environment vars from the file .envrc
include .envrc

# --------------------------------------------------------------------------- #
# HELPERS
# --------------------------------------------------------------------------- #

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N]' && read ans && [ $${ans:-N} = y ]

# --------------------------------------------------------------------------- #
# DEVELOPMENT
# --------------------------------------------------------------------------- #

## run/api: run the ./cmd/api application
.PHONY:
run/api:
	go run ./cmd/api -db-dsn=${GREENLIGHT_DB_DSN}

## db/psql: connect to the DB with psql
.PHONY: db/psql
db/psql:
	psql ${GREENLIGHT_DB_DSN}

## db/migrations/new: create migrations for files with given names
.PHONY: db/migrations/new
db/migrations/new:
	@echo 'Creating migrations for files ${name}'
	migrate -seq -ext=.sql -dir=./migrations ${name}

## db/migrations/up: migrate up all the migration files in ./migrations
.PHONY: db/migrations/up
db/migrations/up: confirm
	@echo 'Running up migrations...'
	migrate -path ./migrations -database ${GREENLIGHT_DB_DSN} up

# --------------------------------------------------------------------------- #
# QUALITY CONTROL
# --------------------------------------------------------------------------- #

## audit: tidy and dependencies, format, vet, test, and staticcheck
.PHONY: audit
audit: vendor
	@echo 'Formatting code...'
	go fmt ./...
	@echo 'Vetting code...'
	go vet ./...
	${GOBINPATH}/staticcheck ./...
	@echo 'Running test...'
	go test -race -vet=off ./...


## vendor: tidy and vendor dependencies
.PHONY: vendor
vendor:
	@echo 'Tidying and verifying module dependencies...'
	go mod tidy
	go mod verify
	@echo 'Vendoring all dependencies...'
	go mod vendor

# --------------------------------------------------------------------------- #
# BUILD
# --------------------------------------------------------------------------- #

## build: build the cmd/api application
.PHONY: build/api
build/api:
	@echo 'Build cmd/api...'
	go build -ldflags='-s' -o=./bin/api ./cmd/api
	GOOS=linux GOARCH=amd64 go build -ldflags='-s' -o=./bin/linux_amd64/api ./cmd/api
