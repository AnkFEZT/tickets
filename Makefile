.PHONY: lint docker-up docker-down test test-verbose dev create-show book-tickets

# Only include .env.test when running test targets
ifneq (,$(filter test test-verbose,$(MAKECMDGOALS)))
include .env.test
endif

# Only include .env.local when running dev target
ifneq (,$(filter dev,$(MAKECMDGOALS)))
include .env.local
endif

lint:
	@grep -rnw --include="*.go" 'collect \*assert\.CollectT' . && exit 1 || exit 0

docker-down:
	docker compose down -v

docker-up:
	docker compose up -d

test:
	go test ./...

test-verbose:
	go test -v ./...

dev:
	go run cmd/server/main.go