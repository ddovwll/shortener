PKG             ?= ./...
INTEGRATION_PKG ?= ./src/internal/infrastructure/data/integration_tests

up:
	docker-compose up

swagger:
	swag init -g src/cmd/main.go -o src/internal/web_api/docs

test:
	go test $(PKG)

lint:
	golangci-lint run --fix